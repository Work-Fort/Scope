use serde::Deserialize;
use std::collections::HashMap;
use std::path::PathBuf;

use crate::domain::{Fort, ServiceConfig};

#[derive(Debug, Deserialize)]
pub struct Config {
    #[serde(default = "default_listen")]
    pub listen: String,
    #[serde(default = "default_database")]
    pub database: String,
    #[serde(default)]
    pub forts: HashMap<String, FortYaml>,
}

#[derive(Debug, Deserialize)]
pub struct FortYaml {
    #[serde(default = "default_true")]
    pub local: bool,
    pub gateway: Option<String>,
    #[serde(default)]
    pub services: Vec<ServiceYaml>,
}

#[derive(Debug, Deserialize)]
pub struct ServiceYaml {
    pub url: String,
}

fn default_true() -> bool {
    true
}

fn default_listen() -> String {
    "127.0.0.1:16100".into()
}

fn default_database() -> String {
    data_dir().join("scope.db").to_string_lossy().into_owned()
}

/// XDG config directory: $XDG_CONFIG_HOME/workfort or ~/.config/workfort
pub fn config_dir() -> PathBuf {
    directories::ProjectDirs::from("", "", "workfort")
        .map(|d| d.config_dir().to_path_buf())
        .unwrap_or_else(|| PathBuf::from("."))
}

/// XDG data directory: $XDG_DATA_HOME/workfort or ~/.local/share/workfort
pub fn data_dir() -> PathBuf {
    directories::ProjectDirs::from("", "", "workfort")
        .map(|d| d.data_dir().to_path_buf())
        .unwrap_or_else(|| PathBuf::from("."))
}

/// XDG state directory: $XDG_STATE_HOME/workfort or ~/.local/state/workfort
pub fn state_dir() -> PathBuf {
    directories::ProjectDirs::from("", "", "workfort")
        .map(|d| d.state_dir().unwrap_or_else(|| d.data_dir()).to_path_buf())
        .unwrap_or_else(|| PathBuf::from("."))
}

impl Config {
    /// Load config from XDG path. Returns defaults if file doesn't exist.
    pub fn load() -> Result<Self, crate::Error> {
        let path = config_dir().join("config.yaml");
        if !path.exists() {
            return Ok(Self {
                listen: default_listen(),
                database: default_database(),
                forts: HashMap::new(),
            });
        }
        let contents = std::fs::read_to_string(&path)
            .map_err(|e| crate::Error::Config(format!("read {}: {e}", path.display())))?;
        serde_yaml::from_str(&contents)
            .map_err(|e| crate::Error::Config(format!("parse {}: {e}", path.display())))
    }

    /// Load config from a specific path (for testing or CLI override).
    pub fn load_from(path: &str) -> Result<Self, crate::Error> {
        let contents = std::fs::read_to_string(path)
            .map_err(|e| crate::Error::Config(format!("read {path}: {e}")))?;
        serde_yaml::from_str(&contents)
            .map_err(|e| crate::Error::Config(format!("parse {path}: {e}")))
    }

    /// Convert parsed config into domain Fort objects.
    pub fn into_forts(self) -> Vec<Fort> {
        self.forts
            .into_iter()
            .map(|(name, f)| Fort {
                name,
                local: f.local,
                gateway: f.gateway,
                services: f
                    .services
                    .into_iter()
                    .map(|s| ServiceConfig { url: s.url })
                    .collect(),
            })
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    #[test]
    fn parse_full_config() {
        let yaml = r#"
listen: "0.0.0.0:8080"
database: "/tmp/test.db"
forts:
  local:
    local: true
    services:
      - url: "http://localhost:16000"
      - url: "http://localhost:3000"
  remote:
    local: false
    gateway: "https://acme.workfort.dev"
"#;
        let mut tmp = tempfile::NamedTempFile::new().unwrap();
        write!(tmp, "{}", yaml).unwrap();

        let config = Config::load_from(tmp.path().to_str().unwrap()).unwrap();
        assert_eq!(config.listen, "0.0.0.0:8080");
        assert_eq!(config.database, "/tmp/test.db");

        let forts = config.into_forts();
        assert_eq!(forts.len(), 2);

        let local = forts.iter().find(|f| f.name == "local").unwrap();
        assert!(local.local);
        assert_eq!(local.services.len(), 2);

        let remote = forts.iter().find(|f| f.name == "remote").unwrap();
        assert!(!remote.local);
        assert_eq!(remote.gateway.as_deref(), Some("https://acme.workfort.dev"));
    }

    #[test]
    fn defaults_when_empty() {
        let yaml = "";
        let mut tmp = tempfile::NamedTempFile::new().unwrap();
        write!(tmp, "{}", yaml).unwrap();

        let config = Config::load_from(tmp.path().to_str().unwrap()).unwrap();
        assert_eq!(config.listen, "127.0.0.1:16100");
        assert!(config.forts.is_empty());
    }

    #[test]
    fn local_defaults_to_true() {
        let yaml = r#"
forts:
  myfort:
    services:
      - url: "http://localhost:16000"
"#;
        let mut tmp = tempfile::NamedTempFile::new().unwrap();
        write!(tmp, "{}", yaml).unwrap();

        let config = Config::load_from(tmp.path().to_str().unwrap()).unwrap();
        let forts = config.into_forts();
        assert!(forts[0].local);
    }
}
