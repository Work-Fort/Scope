pub mod config;
pub mod domain;
pub mod infra;

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("not found: {0}")]
    NotFound(String),
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("config error: {0}")]
    Config(String),
    #[error("{0}")]
    Other(String),
}
