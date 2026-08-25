ALTER TABLE debatesApp.debates
DROP CONSTRAINT fk_debates_winner_side;

ALTER TABLE debatesApp.debates
DROP COLUMN winner_side_id;