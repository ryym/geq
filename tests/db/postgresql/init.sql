DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id serial NOT NULL,
  name varchar(128) NOT NULL
);
INSERT INTO users VALUES (1, 'user1'), (2, 'user2'), (3, 'user3');

DROP TABLE IF EXISTS posts;
CREATE TABLE posts (
  id serial NOT NULL,
  author_id int NOT NULL,
  title varchar(128) NOT NULL
);
INSERT INTO posts (id, author_id, title) VALUES
  (1, 1, 'user1-post1'),
  (2, 1, 'user1-post2'),
  (3, 2, 'user2-post1'),
  (4, 3, 'user3-post1'),
  (5, 3, 'user3-post2'),
  (6, 3, 'user3-post3');

DROP TABLE IF EXISTS transactions;
CREATE TABLE IF NOT EXISTS transactions
(
  id serial NOT NULL,
  user_id int NOT NULL,
  amount int NOT NULL,
  description varchar(256) NOT NULL
);
