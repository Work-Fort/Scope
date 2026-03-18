use async_trait::async_trait;

use super::{Fort, Notification, NotificationLevel};

pub type Result<T> = std::result::Result<T, crate::Error>;

#[async_trait]
pub trait Store: Send + Sync {
    // Fort config
    async fn list_forts(&self) -> Result<Vec<Fort>>;
    async fn get_fort(&self, name: &str) -> Result<Fort>;
    async fn upsert_fort(&self, fort: &Fort) -> Result<()>;
    async fn delete_fort(&self, name: &str) -> Result<()>;
    async fn get_active_fort(&self) -> Result<Option<String>>;
    async fn set_active_fort(&self, name: &str) -> Result<()>;

    // Notifications
    async fn insert_notification(&self, n: &Notification) -> Result<i64>;
    async fn list_notifications(
        &self,
        limit: i64,
        before_id: Option<i64>,
    ) -> Result<Vec<Notification>>;
    async fn unread_count(&self) -> Result<i64>;
    async fn mark_read(&self, up_to_id: i64) -> Result<()>;

    // User preferences
    async fn get_preference(&self, service: &str) -> Result<NotificationLevel>;
    async fn set_preference(&self, service: &str, level: NotificationLevel) -> Result<()>;
}
