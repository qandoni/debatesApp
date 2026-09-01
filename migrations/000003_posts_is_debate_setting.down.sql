ALTER TABLE debatesApp.posts
    ALTER COLUMN is_debate DROP NOT NULL,
    ALTER COLUMN is_debate DROP DEFAULT;