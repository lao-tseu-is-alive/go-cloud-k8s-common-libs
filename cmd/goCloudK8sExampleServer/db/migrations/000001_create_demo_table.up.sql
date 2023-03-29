create table if not exists demo
(
    id          serial constraint demo_pk primary key,
    name        text not null constraint demo_unique_name unique,
    description text,
    created_at  timestamp default now() not null
);

comment on table demo is 'demo table to illustrate using db golang-migrate in project ';
comment on column demo.name is 'name of a demo entry';
comment on constraint demo_unique_name on demo is 'a name for the demo should be unique';

