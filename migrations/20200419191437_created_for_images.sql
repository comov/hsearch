-- +goose Up
-- +goose StatementBegin
alter table image
    add created integer default 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
create table image_dg_tmp
(
    offer_id integer not null,
    path varchar(255) default ''
);

insert into image_dg_tmp(offer_id, path) select offer_id, path from image;

drop table image;

alter table image_dg_tmp rename to image;

create index image_offer_id_index
    on image (offer_id);

create unique index image_path_uindex
    on image (path);

-- +goose StatementEnd
