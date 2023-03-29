create table if not exists demo_type
(
    id          serial constraint demo_type_pk primary key,
    name        text not null constraint demo_type_unique_name unique,
    created_at  timestamp default now() not null
);

comment on table demo_type is 'demo_type table to store type of demo';
comment on constraint demo_type_unique_name on demo_type is 'a name for the demo_type should be unique';

INSERT INTO demo_type (name) VALUES ('main');
INSERT INTO demo_type (name) VALUES ('secondary');

ALTER TABLE demo ADD COLUMN id_demo_type integer;

alter table demo add constraint fk_demo_demo_type foreign key (id_demo_type) references demo_type (id);
