create table feedback
(
    id       serial  not null
        constraint feedback_pk primary key,
    created  integer not null,
    chat     bigint  not null,
    username varchar(100) default '',
    body     text         default ''
);

create unique index feedback_id_uindex
    on feedback (id);

---- create above / drop below ----
drop table if exists feedback;
