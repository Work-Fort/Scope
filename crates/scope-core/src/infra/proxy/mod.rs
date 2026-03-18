use futures_util::{SinkExt, StreamExt};
use reqwest::Client;
use std::time::Duration;
use tokio_tungstenite::tungstenite::Message;

pub struct ProxyHandler {
    client: Client,
}

pub struct ProxyResponse {
    pub status: u16,
    pub headers: Vec<(String, String)>,
    pub body: Vec<u8>,
}

impl ProxyHandler {
    pub fn new() -> Self {
        let client = Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .expect("failed to build HTTP client");
        Self { client }
    }

    /// Forward an HTTP request to a backend service.
    /// Copies method, path, query, headers, and body.
    /// Attaches Authorization header if token is provided.
    pub async fn forward_http(
        &self,
        service_url: &str,
        method: &str,
        path: &str,
        query: Option<&str>,
        headers: &[(String, String)],
        body: Option<Vec<u8>>,
        token: Option<&str>,
    ) -> Result<ProxyResponse, crate::Error> {
        // Build target URL
        let base = service_url.trim_end_matches('/');
        let path = if path.starts_with('/') {
            path.to_string()
        } else {
            format!("/{path}")
        };
        let target = match query {
            Some(q) if !q.is_empty() => format!("{base}{path}?{q}"),
            _ => format!("{base}{path}"),
        };

        let method = method
            .parse::<reqwest::Method>()
            .map_err(|e| crate::Error::Other(format!("invalid HTTP method: {e}")))?;

        let mut req = self.client.request(method, &target);

        // Rewrite headers for proxying. The BFF is the trust boundary —
        // backend services see scope-server as the client, not the browser.
        //
        // Host: reqwest sets from target URL.
        // Origin: rewritten to the target service's origin (required by Better Auth CSRF).
        // Referer: stripped (leaks browser URL, not useful for backend).
        let target_origin = {
            let u = url::Url::parse(&target).unwrap_or_else(|_| url::Url::parse(service_url).unwrap());
            format!("{}://{}", u.scheme(), u.host_str().unwrap_or("localhost"))
                + &u.port().map(|p| format!(":{p}")).unwrap_or_default()
        };

        let mut has_origin = false;
        for (name, value) in headers {
            if name.eq_ignore_ascii_case("host")
                || name.eq_ignore_ascii_case("referer")
                || name.starts_with("sec-")
            {
                continue;
            }
            if name.eq_ignore_ascii_case("origin") {
                log::debug!("proxy: rewriting Origin from '{}' to '{}'", value, target_origin);
                has_origin = true;
                req = req.header("origin", &target_origin);
                continue;
            }
            // Forward session token cookie but strip session_data (contains origin info
            // that causes CSRF rejection). Other cookies are also stripped.
            if name.eq_ignore_ascii_case("cookie") {
                let filtered: Vec<&str> = value.split(';')
                    .map(|c| c.trim())
                    .filter(|c| c.starts_with("better-auth.session_token="))
                    .collect();
                if !filtered.is_empty() {
                    req = req.header("cookie", filtered.join("; "));
                }
                continue;
            }
            req = req.header(name.as_str(), value.as_str());
        }
        // If browser didn't send Origin (e.g. same-origin navigation), add it
        // for backend services that require it (like Better Auth CSRF).
        if !has_origin {
            log::debug!("proxy: adding Origin header: {}", target_origin);
            req = req.header("origin", &target_origin);
        }

        // Debug: log all outgoing headers
        if let Some(built) = req.try_clone() {
            let built = built.build().ok();
            if let Some(r) = built {
                for (k, v) in r.headers() {
                    log::debug!("proxy outgoing: {} = {}", k, v.to_str().unwrap_or("?"));
                }
            }
        }

        // Attach auth token
        if let Some(tok) = token {
            req = req.header("Authorization", format!("Bearer {tok}"));
        }

        // Attach body
        if let Some(b) = body {
            req = req.body(b);
        }

        let resp = req
            .send()
            .await
            .map_err(|e| crate::Error::Other(format!("proxy request failed: {e}")))?;

        let status = resp.status().as_u16();
        let resp_headers: Vec<(String, String)> = resp
            .headers()
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_str().unwrap_or("").to_string()))
            .collect();
        let resp_body = resp
            .bytes()
            .await
            .map_err(|e| crate::Error::Other(format!("proxy response read failed: {e}")))?
            .to_vec();

        Ok(ProxyResponse {
            status,
            headers: resp_headers,
            body: resp_body,
        })
    }
}

impl Default for ProxyHandler {
    fn default() -> Self {
        Self::new()
    }
}

/// Pipe messages between a client WebSocket and a backend service WebSocket.
/// Returns when either side closes.
pub async fn pipe_ws(
    service_ws_url: &str,
    token: Option<&str>,
    mut client_rx: impl futures_util::Stream<Item = Result<Message, tokio_tungstenite::tungstenite::Error>>
        + Unpin,
    mut client_tx: impl futures_util::Sink<Message, Error = tokio_tungstenite::tungstenite::Error>
        + Unpin,
) -> Result<(), crate::Error> {
    use tokio_tungstenite::tungstenite::client::IntoClientRequest;

    let mut request = service_ws_url
        .into_client_request()
        .map_err(|e| crate::Error::Other(format!("invalid WS URL: {e}")))?;

    if let Some(tok) = token {
        request.headers_mut().insert(
            "Authorization",
            format!("Bearer {tok}")
                .parse()
                .map_err(|e| crate::Error::Other(format!("invalid auth header: {e}")))?,
        );
    }

    let (backend_ws, _resp) = tokio_tungstenite::connect_async(request)
        .await
        .map_err(|e| crate::Error::Other(format!("backend WS connect failed: {e}")))?;

    let (mut backend_tx, mut backend_rx) = backend_ws.split();

    // client → backend
    let c2b = async {
        while let Some(msg) = client_rx.next().await {
            match msg {
                Ok(m) => {
                    if backend_tx.send(m).await.is_err() {
                        break;
                    }
                }
                Err(_) => break,
            }
        }
        let _ = backend_tx.close().await;
    };

    // backend → client
    let b2c = async {
        while let Some(msg) = backend_rx.next().await {
            match msg {
                Ok(m) => {
                    if client_tx.send(m).await.is_err() {
                        break;
                    }
                }
                Err(_) => break,
            }
        }
        let _ = client_tx.close().await;
    };

    // Run both directions concurrently; when either finishes, the other
    // will see a closed channel on its next send/recv and exit.
    tokio::select! {
        _ = c2b => {},
        _ = b2c => {},
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{extract::Request, routing::get, Router};
    use std::net::SocketAddr;
    use tokio::net::TcpListener;

    /// Start a simple axum server on a random port, returning the base URL.
    async fn start_echo_server() -> String {
        let app = Router::new()
            .route(
                "/echo",
                get(|| async { "hello from echo" }).post(
                    |req: Request| async move {
                        let body = axum::body::to_bytes(req.into_body(), 1024 * 1024)
                            .await
                            .unwrap();
                        String::from_utf8(body.to_vec()).unwrap()
                    },
                ),
            )
            .route(
                "/auth-check",
                get(|req: Request| async move {
                    req.headers()
                        .get("authorization")
                        .map(|v| v.to_str().unwrap().to_string())
                        .unwrap_or_else(|| "no-auth".to_string())
                }),
            );

        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr: SocketAddr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });
        format!("http://{addr}")
    }

    #[tokio::test]
    async fn proxy_forwards_get() {
        let base = start_echo_server().await;
        let proxy = ProxyHandler::new();

        let resp = proxy
            .forward_http(&base, "GET", "/echo", None, &[], None, None)
            .await
            .unwrap();

        assert_eq!(resp.status, 200);
        assert_eq!(String::from_utf8(resp.body).unwrap(), "hello from echo");
    }

    #[tokio::test]
    async fn proxy_forwards_post_body() {
        let base = start_echo_server().await;
        let proxy = ProxyHandler::new();

        let resp = proxy
            .forward_http(
                &base,
                "POST",
                "/echo",
                None,
                &[],
                Some(b"request body here".to_vec()),
                None,
            )
            .await
            .unwrap();

        assert_eq!(resp.status, 200);
        assert_eq!(
            String::from_utf8(resp.body).unwrap(),
            "request body here"
        );
    }

    #[tokio::test]
    async fn proxy_attaches_auth_header() {
        let base = start_echo_server().await;
        let proxy = ProxyHandler::new();

        let resp = proxy
            .forward_http(
                &base,
                "GET",
                "/auth-check",
                None,
                &[],
                None,
                Some("my-secret-token"),
            )
            .await
            .unwrap();

        assert_eq!(resp.status, 200);
        assert_eq!(
            String::from_utf8(resp.body).unwrap(),
            "Bearer my-secret-token"
        );
    }

    #[tokio::test]
    async fn proxy_forwards_query_string() {
        let app = Router::new().route(
            "/qs",
            get(|req: Request| async move {
                req.uri().query().unwrap_or("none").to_string()
            }),
        );

        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });
        let base = format!("http://{addr}");

        let proxy = ProxyHandler::new();
        let resp = proxy
            .forward_http(&base, "GET", "/qs", Some("foo=bar&x=1"), &[], None, None)
            .await
            .unwrap();

        assert_eq!(resp.status, 200);
        assert_eq!(String::from_utf8(resp.body).unwrap(), "foo=bar&x=1");
    }

    #[tokio::test]
    async fn proxy_skips_host_header() {
        let app = Router::new().route(
            "/host-check",
            get(|req: Request| async move {
                req.headers()
                    .get("host")
                    .map(|v| v.to_str().unwrap().to_string())
                    .unwrap_or_else(|| "no-host".to_string())
            }),
        );

        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });
        let base = format!("http://{addr}");

        let proxy = ProxyHandler::new();
        let resp = proxy
            .forward_http(
                &base,
                "GET",
                "/host-check",
                None,
                &[("Host".to_string(), "evil.example.com".to_string())],
                None,
                None,
            )
            .await
            .unwrap();

        assert_eq!(resp.status, 200);
        // The Host header should be set by reqwest to the target, not the custom one
        let body = String::from_utf8(resp.body).unwrap();
        assert!(!body.contains("evil.example.com"), "Host header should be overridden, got: {body}");
    }
}
