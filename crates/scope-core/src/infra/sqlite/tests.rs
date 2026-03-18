use chrono::Utc;

use crate::domain::{
    Fort, Notification, NotificationLevel, ServiceConfig, Store, Urgency,
};

use super::SqliteStore;

async fn mem_store() -> SqliteStore {
    SqliteStore::open(":memory:").await.unwrap()
}

#[tokio::test]
async fn fort_crud() {
    let store = mem_store().await;

    // Initially empty
    let forts = store.list_forts().await.unwrap();
    assert!(forts.is_empty());

    // Insert
    let fort = Fort {
        name: "local".into(),
        local: true,
        gateway: None,
        services: vec![
            ServiceConfig { url: "http://localhost:8080".into() },
            ServiceConfig { url: "http://localhost:9090".into() },
        ],
    };
    store.upsert_fort(&fort).await.unwrap();

    // List
    let forts = store.list_forts().await.unwrap();
    assert_eq!(forts.len(), 1);
    assert_eq!(forts[0].name, "local");
    assert!(forts[0].local);
    assert_eq!(forts[0].services.len(), 2);

    // Get
    let f = store.get_fort("local").await.unwrap();
    assert_eq!(f.name, "local");
    assert_eq!(f.services.len(), 2);

    // Update — change gateway and services
    let updated = Fort {
        name: "local".into(),
        local: true,
        gateway: Some("https://gw.example.com".into()),
        services: vec![ServiceConfig { url: "http://localhost:3000".into() }],
    };
    store.upsert_fort(&updated).await.unwrap();
    let f = store.get_fort("local").await.unwrap();
    assert_eq!(f.gateway.as_deref(), Some("https://gw.example.com"));
    assert_eq!(f.services.len(), 1);
    assert_eq!(f.services[0].url, "http://localhost:3000");

    // Delete
    store.delete_fort("local").await.unwrap();
    let forts = store.list_forts().await.unwrap();
    assert!(forts.is_empty());

    // Get missing → NotFound
    let err = store.get_fort("local").await.unwrap_err();
    assert!(matches!(err, crate::Error::NotFound(_)));
}

#[tokio::test]
async fn active_fort() {
    let store = mem_store().await;

    store
        .upsert_fort(&Fort {
            name: "a".into(),
            local: true,
            gateway: None,
            services: vec![],
        })
        .await
        .unwrap();
    store
        .upsert_fort(&Fort {
            name: "b".into(),
            local: false,
            gateway: Some("https://b.example.com".into()),
            services: vec![],
        })
        .await
        .unwrap();

    // No active fort initially
    assert!(store.get_active_fort().await.unwrap().is_none());

    // Set active
    store.set_active_fort("a").await.unwrap();
    assert_eq!(store.get_active_fort().await.unwrap().as_deref(), Some("a"));

    // Switch active
    store.set_active_fort("b").await.unwrap();
    assert_eq!(store.get_active_fort().await.unwrap().as_deref(), Some("b"));

    // Setting active to non-existent fort → NotFound
    let err = store.set_active_fort("nope").await.unwrap_err();
    assert!(matches!(err, crate::Error::NotFound(_)));
}

#[tokio::test]
async fn fort_delete_cascades_services() {
    let store = mem_store().await;

    store
        .upsert_fort(&Fort {
            name: "x".into(),
            local: true,
            gateway: None,
            services: vec![
                ServiceConfig { url: "http://a".into() },
                ServiceConfig { url: "http://b".into() },
            ],
        })
        .await
        .unwrap();

    // Verify services exist
    let f = store.get_fort("x").await.unwrap();
    assert_eq!(f.services.len(), 2);

    // Delete fort — services should cascade
    store.delete_fort("x").await.unwrap();

    // Verify no orphaned services via raw query
    let rows = sqlx::query("SELECT COUNT(*) as cnt FROM fort_services WHERE fort_name = 'x'")
        .fetch_one(&store.pool)
        .await
        .unwrap();
    let cnt: i64 = sqlx::Row::get(&rows, "cnt");
    assert_eq!(cnt, 0);
}

#[tokio::test]
async fn notification_crud() {
    let store = mem_store().await;

    // Insert several notifications
    let now = Utc::now();
    let mut ids = Vec::new();
    for i in 0..5 {
        let id = store
            .insert_notification(&Notification {
                id: None,
                service: "chat".into(),
                title: format!("msg {i}"),
                body: Some(format!("body {i}")),
                urgency: if i % 2 == 0 { Urgency::Passive } else { Urgency::Active },
                route: Some("/chat".into()),
                read: false,
                created_at: now,
            })
            .await
            .unwrap();
        ids.push(id);
    }

    // Unread count
    assert_eq!(store.unread_count().await.unwrap(), 5);

    // List all (limit 10) — newest first
    let all = store.list_notifications(10, None).await.unwrap();
    assert_eq!(all.len(), 5);
    assert_eq!(all[0].title, "msg 4");
    assert_eq!(all[4].title, "msg 0");

    // Pagination with before_id
    let page = store.list_notifications(2, Some(ids[4])).await.unwrap();
    assert_eq!(page.len(), 2);
    assert_eq!(page[0].title, "msg 3");
    assert_eq!(page[1].title, "msg 2");

    // Mark read up to id 3
    store.mark_read(ids[2]).await.unwrap();
    assert_eq!(store.unread_count().await.unwrap(), 2);

    // Verify read flag
    let all = store.list_notifications(10, None).await.unwrap();
    assert!(!all[0].read); // msg 4
    assert!(!all[1].read); // msg 3
    assert!(all[2].read);  // msg 2
    assert!(all[3].read);  // msg 1
    assert!(all[4].read);  // msg 0
}

#[tokio::test]
async fn preferences() {
    let store = mem_store().await;

    // Default when no row
    let level = store.get_preference("chat").await.unwrap();
    assert_eq!(level, NotificationLevel::AllowUrgent);

    // Set
    store
        .set_preference("chat", NotificationLevel::Mute)
        .await
        .unwrap();
    assert_eq!(
        store.get_preference("chat").await.unwrap(),
        NotificationLevel::Mute,
    );

    // Update
    store
        .set_preference("chat", NotificationLevel::PassiveOnly)
        .await
        .unwrap();
    assert_eq!(
        store.get_preference("chat").await.unwrap(),
        NotificationLevel::PassiveOnly,
    );

    // Different service still has default
    assert_eq!(
        store.get_preference("other").await.unwrap(),
        NotificationLevel::AllowUrgent,
    );
}
