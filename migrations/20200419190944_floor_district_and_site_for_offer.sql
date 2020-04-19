-- +goose Up
-- +goose StatementBegin
alter table offer
    add site varchar(20) default '';

alter table offer
    add floor varchar(20) default '';

alter table offer
    add district varchar(20) default '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
create table offer_dg_tmp
(
    id integer not null,
    created integer not null,
    url varchar(255) default '',
    topic varchar(255) default '',
    full_price varchar(50) default '',
    phone varchar(255) default '',
    room_numbers varchar(255) default '',
    body text default '',
    images integer default 0 not null,
    price real default 0,
    currency varchar(10) default '',
    area varchar(100) default '',
    city varchar(100) default '',
    room_type varchar(100) default ''
);

insert into offer_dg_tmp(id, created, url, topic, full_price, phone, room_numbers, body, images, price, currency, area, city, room_type) select id, created, url, topic, full_price, phone, room_numbers, body, images, price, currency, area, city, room_type from offer;

drop table offer;

alter table offer_dg_tmp rename to offer;

create unique index offer_id_uindex
    on offer (id);

-- +goose StatementEnd
