-- generated by github.com/mackee/go-genddl. DO NOT EDIT!!!

DROP TABLE IF EXISTS "product";

CREATE TABLE "product" (
    "id" INTEGER NOT NULL AUTOINCREMENT,
    "name" TEXT NOT NULL,
    "type" INTEGER NOT NULL,
    "user_id" INTEGER NOT NULL,
    "created_at" DATETIME NOT NULL,
    PRIMARY KEY ("id", "created_at"),
    UNIQUE ("user_id", "type"),
    INDEX ("user_id", "created_at"),
    FOREIGN KEY ("user_id") REFERENCES user("id") ON DELETE CASCADE ON UPDATE SET CASCADE
);

DROP TABLE IF EXISTS "user";

CREATE TABLE "user" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL UNIQUE,
    "age" INTEGER NULL,
    "message" TEXT NULL,
    "created_at" DATETIME NOT NULL,
    "updated_at" DATETIME NULL
);