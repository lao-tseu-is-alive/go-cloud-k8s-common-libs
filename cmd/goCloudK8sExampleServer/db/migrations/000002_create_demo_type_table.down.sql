ALTER TABLE demo DROP COLUMN id_demo_type;
ALTER TABLE demo DROP CONSTRAINT fk_demo_demo_type;
DROP TABLE IF EXISTS demo_type;


