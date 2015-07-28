INSERT INTO post (title, body) VALUES ('test 1', 'lorem impsum for test 1');
INSERT INTO post (title, body) VALUES ('test 2', 'lorem impsum for test 2');
INSERT INTO post (title, body) VALUES ('test 3', 'lorem impsum for test 3');
INSERT INTO post (title, body) VALUES ('test 4', 'lorem impsum for test 4');

INSERT INTO comments (author, body, post_id, approved) VALUES ('author 1', 'comment 1', 1, 't');
INSERT INTO comments (author, body, post_id, approved) VALUES ('author 2', 'comment 2', 1, 'f');
INSERT INTO comments (author, body, post_id, approved) VALUES ('author 3', 'comment 3', 3, 't');
INSERT INTO comments (author, body, post_id, approved) VALUES ('author 4', 'comment 4', 4, 'f');
