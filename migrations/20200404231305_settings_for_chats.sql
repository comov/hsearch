-- +goose Up
-- +goose StatementBegin
alter table chat
    add created integer default 0;

alter table chat
    add diesel integer default 1;

alter table chat
    add lalafo integer default 1;

alter table chat
    add photo integer default 0;

alter table chat
    add usd varchar(100) default '0:0';

alter table chat
    add kgs varchar(100) default '0:0';

alter table chat
    add up_track integer default 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
create table chat_dg_tmp
(
    id       integer not null,
    username varchar(100) default '',
    enable   integer      default 0,
    title    varchar(100) default '',
    c_type   varchar(20)  default ''
);

insert into chat_dg_tmp(id, username, enable, title, c_type)
select id, username, enable, title, c_type
from chat;

drop table chat;

alter table chat_dg_tmp
    rename to chat;

create unique index chat_chat_index
    on chat (id);

create unique index chat_name_uindex
    on chat (username);

-- +goose StatementEnd
