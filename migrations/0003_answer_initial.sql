create table answer
(
    id       serial  not null
        constraint answer_pk primary key,
    created  integer not null,
    chat     bigint  not null,
    offer_id integer not null,
    dislike  boolean default false
);

create index answer_offer_id_index
    on answer (offer_id);

create index answer_chat_index
    on answer (chat);

create unique index answer_id_uindex
    on answer (id);

---- create above / drop below ----
drop table if exists answer;
