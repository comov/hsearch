-- +goose Up
-- +goose StatementBegin
create table feedback
(
    created  integer not null,
    chat     integer not null,
    username varchar(100) default '',
    body     text         default ''
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists feedback;
-- +goose StatementEnd
