ALTER TABLE debatesApp.debates
ADD COLUMN winner_side_id INT;

ALTER TABLE debatesApp.debates
ADD CONSTRAINT fk_debates_winner_side
FOREIGN KEY (winner_side_id)
REFERENCES debatesApp.debate_sides(id);