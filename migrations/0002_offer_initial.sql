create table offer
(
    id           integer                not null,
    created      integer                not null,
    url          varchar(255) default '',
    topic        varchar(255) default '',
    full_price   varchar(50)  default '',
    phone        varchar(255) default '',
    room_numbers varchar(255) default '',
    body         text         default '',
    images       integer      default 0 not null,
    price        integer      default 0,
    currency     varchar(10)  default '',
    area         varchar(100) default '',
    city         varchar(100) default '',
    room_type    varchar(100) default '',
    site         varchar(20)  default '',
    floor        varchar(20)  default '',
    district     varchar(100)  default ''
);

create unique index offer_id_uindex
    on offer (id);

-- sqlite не умеет хранить списки в полях, по этому делам отдельную таблицу
-- todo: fixed after migrate to PG
create table image
(
    id       serial  not null
        constraint image_pk primary key,
    offer_id integer not null,
    path     varchar(255) default '',
    created  integer      default 0
);

create index image_offer_id_index
    on image (offer_id);

create unique index image_path_uindex
    on image (path);

create unique index image_id_uindex
    on image (id);

---- create above / drop below ----
drop table if exists offer;
drop table if exists image;
