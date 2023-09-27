package tests

const initMySQL = `
DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id int unsigned NOT NULL PRIMARY KEY AUTO_INCREMENT,
  name varchar(128) NOT NULL
);

DROP TABLE IF EXISTS posts;
CREATE TABLE posts (
  id int unsigned NOT NULL PRIMARY KEY AUTO_INCREMENT,
  author_id int unsigned NOT NULL,
  title varchar(128) NOT NULL
);

DROP TABLE IF EXISTS transactions;
CREATE TABLE transactions (
  id int unsigned NOT NULL PRIMARY KEY AUTO_INCREMENT,
  user_id int unsigned NOT NULL,
  amount int NOT NULL,
  description varchar(256) NOT NULL DEFAULT '',
  created_at datetime NOT NULL DEFAULT NOW()
);
`

const initPostgreSQL = `
DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id serial NOT NULL,
  name varchar(128) NOT NULL
);

DROP TABLE IF EXISTS posts;
CREATE TABLE posts (
  id serial NOT NULL,
  author_id int NOT NULL,
  title varchar(128) NOT NULL
);

DROP TABLE IF EXISTS transactions;
CREATE TABLE transactions (
  id serial NOT NULL,
  user_id int NOT NULL,
  amount int NOT NULL,
  description varchar(256) NOT NULL DEFAULT '',
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const fixtureSQL = `
INSERT INTO users VALUES (1, 'user1'), (2, 'user2'), (3, 'user3');
INSERT INTO posts (id, author_id, title) VALUES
  (1, 1, 'user1-post1'),
  (2, 1, 'user1-post2'),
  (3, 2, 'user2-post1'),
  (4, 3, 'user3-post1'),
  (5, 3, 'user3-post2'),
  (6, 3, 'user3-post3');
`
