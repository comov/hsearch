create table tg_messages
(
    id         serial  not null
        constraint tg_messages_pk primary key,
    created    integer not null,
    message_id integer not null,
    offer_id   integer not null,
    chat       bigint  not null,
    kind       varchar(50) default ''
);

create index tg_messages_message_id_index
    on tg_messages (message_id);

create index tg_messages_offer_id_index
    on tg_messages (offer_id);

create index tg_messages_chat_index
    on tg_messages (chat);

create unique index tg_messages_id_uindex
    on tg_messages (id);

---- create above / drop below ----
drop table if exists tg_messages;
