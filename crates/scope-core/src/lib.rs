pub mod config;
pub mod domain;
pub mod infra;

use std::sync::Arc;

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

pub async fn open_store(url: &str) -> Result<Arc<dyn domain::Store>, Error> {
    if url.starts_with("postgres://") || url.starts_with("postgresql://") {
        let store = infra::postgres::PostgresStore::open(url).await?;
        Ok(Arc::new(store))
    } else {
        let store = infra::sqlite::SqliteStore::open(url).await?;
        Ok(Arc::new(store))
    }
}
