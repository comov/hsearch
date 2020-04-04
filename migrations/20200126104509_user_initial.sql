-- +goose Up
-- +goose StatementBegin
create table user
(
    username varchar(100) default '',
    chat     integer not null,
    enable   integer      default 0
);

create unique index user_chat_index
    on user (chat);

create unique index user_name_uindex
    on user (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists user;
-- +goose StatementEnd
