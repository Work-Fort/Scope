// SPA fallback is configured directly in main.rs using tower-http's ServeDir.
//
// Usage:
//   use tower_http::services::{ServeDir, ServeFile};
//   let spa = ServeDir::new("web/shell/dist")
//       .not_found_service(ServeFile::new("web/shell/dist/index.html"));
//   // then: .fallback_service(spa)
