-- +goose Up
-- +goose StatementBegin
alter table user
    rename to chat;

create table chat_dg_tmp
(
    id       integer not null,
    username varchar(100) default '',
    enable   integer      default 0,
    title    varchar(100) default '',
    c_type   varchar(20)  default ''
);

insert into chat_dg_tmp(username, id, enable)
select username, chat, enable
from chat;

drop table chat;

alter table chat_dg_tmp
    rename to chat;

create unique index chat_chat_index
    on chat (id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table chat
    rename to user;

create table chat_dg_tmp
(
    chat     integer not null,
    username varchar(100) default '',
    enable   integer      default 0
);

insert into user_dg_tmp(chat, username, enable)
select id, username, enable
from user;

drop table user;

alter table user_dg_tmp
    rename to user;

create unique index user_chat_index
    on user (chat);

create unique index user_name_uindex
    on user (username);

-- +goose StatementEnd
