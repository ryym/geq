CREATE DATABASE IF NOT EXISTS geq CHARACTER SET utf8mb4 collate utf8mb4_bin;

USE geq;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL
);
INSERT INTO users VALUES (1, "user1"), (2, "user2"), (3, "user3");

DROP TABLE IF EXISTS posts;
CREATE TABLE posts (
  id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  author_id INT UNSIGNED NOT NULL,
  title VARCHAR(128) NOT NULL
);
INSERT INTO posts (id, author_id, title) VALUES
  (1, 1, "user1-post1"),
  (2, 1, "user1-post2"),
  (3, 2, "user2-post1"),
  (4, 3, "user3-post1"),
  (5, 3, "user3-post2"),
  (6, 3, "user3-post3");
