module.exports = {
    async up(db, client) {
        await db.createCollection("users");
        await db.createCollection("categories");
        await db.createCollection("datasets");
        await db.createCollection("dataset_versions");
        await db.createCollection("metadata");
        await db.createCollection("subscriptions");
        await db.createCollection("reviews");
        await db.createCollection("notifications");
        await db.createCollection("access_requests");

        await db.collection("users").createIndex({ email: 1 }, { unique: true });
        await db.collection("users").createIndex({ registration_date: -1 });

        await db.collection("categories").createIndex({ name: 1 }, { unique: true });

        await db.collection("datasets").createIndex({ owner_id: 1 });
        await db.collection("datasets").createIndex(
            { category_id: 1, is_public: 1, created_at: -1 }
        );

        await db.collection("dataset_versions").createIndex(
            { dataset_id: 1, upload_date: -1 }
        );

        await db.collection("metadata").createIndex(
            { dataset_version_id: 1 }
        );

        await db.collection("subscriptions").createIndex(
            { user_id: 1, dataset_id: 1 },
            { unique: true }
        );
        await db.collection("subscriptions").createIndex(
            { dataset_id: 1 }
        );

        await db.collection("reviews").createIndex(
            { dataset_id: 1, created_at: -1 }
        );
        await db.collection("reviews").createIndex(
            { user_id: 1, created_at: -1 }
        );

        await db.collection("notifications").createIndex(
            { user_id: 1, created_at: -1 }
        );

        await db.collection("access_requests").createIndex(
            { dataset_id: 1, user_id: 1 },
            { unique: true }
        );
        await db.collection("access_requests").createIndex(
            { created_at: -1 }
        );

        const names = [
            "users", "categories", "datasets", "dataset_versions",
            "metadata", "reviews", "notifications", "access_requests"
        ];
        const counters = names.map(n => ({ _id: n, seq: 0 }));
        await db.collection("counters").insertMany(counters);
    },

    async down(db, client) {
        const colls = [
            "access_requests", "notifications", "reviews", "subscriptions",
            "metadata", "dataset_versions", "datasets", "categories", "users", "counters"
        ];
        for (const c of colls) {
            try { await db.collection(c).drop(); }
            catch (e) {  }
        }
    }
};
