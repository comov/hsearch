-- +goose Up
-- +goose StatementBegin
alter table tg_messages
    add kind varchar(50) default '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- https://www.techonthenet.com/sqlite/tables/alter_table.php
create table tg_messages_dg_tmp
(
    created    integer not null,
    message_id integer not null,
    offer_id   integer not null,
    chat       integer not null
);

insert into tg_messages_dg_tmp(created, message_id, offer_id, chat)
select created, message_id, offer_id, chat
from tg_messages;

drop table tg_messages;

alter table tg_messages_dg_tmp
    rename to tg_messages;

create index tg_messages_chat_index
    on tg_messages (chat);

create index tg_messages_message_id_index
    on tg_messages (message_id);

create unique index tg_messages_message_id_offer_id_chat_uindex
    on tg_messages (message_id, offer_id, chat);

create index tg_messages_offer_id_index
    on tg_messages (offer_id);
-- +goose StatementEnd
