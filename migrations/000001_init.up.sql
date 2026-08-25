CREATE SCHEMA debatesApp;

CREATE TABLE debatesApp.users(
    id              SERIAL                  PRIMARY KEY,
    version         INT         NOT NULL    DEFAULT 1,
    username        TEXT        NOT NULL,
    email           TEXT        NOT NULL    UNIQUE,
    password_hash   TEXT        NOT NULL,
    avatar_url      TEXT,
    bio             TEXT,
    created_at      TIMESTAMPTZ NOT NULL    DEFAULT NOW(),
    updated_at      TIMESTAMPTZ 
);

CREATE TABLE debatesApp.posts(
    id              SERIAL                  PRIMARY KEY,
    version         INT         NOT NULL    DEFAULT 1,
    author_id       INT         NOT NULL    REFERENCES debatesApp.users(id),
    content         TEXT,
    is_debate       BOOLEAN,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE debatesApp.debates(
    id              SERIAL                  PRIMARY KEY,
    version         INT         NOT NULL    DEFAULT 1,
    post_id         INT         NOT NULL    REFERENCES debatesApp.posts(id),
    status          TEXT        NOT NULL    CHECK(status IN(
        'open',
        'completed'
    )),
    end_at          TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL    DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);

CREATE TABLE debatesApp.debate_sides(
    id              SERIAL      NOT NULL PRIMARY KEY,
    version         INT         NOT NULL DEFAULT 1,
    debate_id       INT         NOT NULL REFERENCES debatesApp.debates(id),
    name            TEXT        NOT NULL,
    description     TEXT,
    display_order   INT
);

CREATE TABLE debatesApp.debate_votes(
    id              SERIAL      NOT NULL PRIMARY KEY,
    version         INT         NOT NULL DEFAULT 1,
    debate_id       INT         NOT NULL REFERENCES debatesApp.debates(id),
    user_id         INT         NOT NULL REFERENCES debatesApp.users(id),
    debate_side_id  INT         NOT NULL REFERENCES debatesApp.debate_sides(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ,
    is_changed      BOOLEAN
);

CREATE TABLE debatesApp.post_images(
    id              SERIAL      NOT NULL PRIMARY KEY,
    post_id         INT         NOT NULL REFERENCES debatesApp.posts(id),
    image_url       TEXT        NOT NULL,
    display_order   INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE debatesApp.comments(
    id                  SERIAL      NOT NULL       PRIMARY KEY,
    version             INT         NOT NULL       DEFAULT 1, 
    post_id             INT         NOT NULL       REFERENCES debatesApp.posts(id),
    parent_comment_id   INT                        REFERENCES debatesApp.comments(id),
    author_id           INT         NOT NULL       REFERENCES debatesApp.users(id),
    debate_side_id      INT                        REFERENCES debatesApp.debate_sides(id),
    content             TEXT,
    created_at          TIMESTAMPTZ NOT NULL       DEFAULT NOW(),
    updated_at          TIMESTAMPTZ
);

CREATE TABLE debatesApp.comment_ratings(
    id                  SERIAL      NOT NULL        PRIMARY KEY,
    version             INT         NOT NULL        DEFAULT 1,
    comment_id          INT         NOT NULL UNIQUE REFERENCES debatesApp.comments(id),
    user_id             INT         NOT NULL UNIQUE REFERENCES debatesApp.users(id),
    score               INT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ
);