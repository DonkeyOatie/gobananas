CREATE TABLE post (
  id BIGSERIAL PRIMARY KEY,
  title varchar(30) NOT NULL,
  body text NOT NULL,
  date_published timestamp NOT NULL DEFAULT now()
);

CREATE TABLE comments (
  author varchar(50),
  body text,
  date_commented timestamp NOT NULL DEFAULT now(),
  post_id integer REFERENCES post (id) ON DELETE CASCADE,
  approved boolean DEFAULT FALSE
)
