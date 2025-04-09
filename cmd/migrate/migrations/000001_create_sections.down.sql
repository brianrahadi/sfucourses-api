select * from schema_migrations;

update schema_migrations set dirty=false where version=1;
DROP TABLE IF EXISTS schema_migrations;
DROP TABLE IF EXISTS sections;
