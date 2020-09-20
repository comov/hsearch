CREATE TABLE IF NOT EXISTS chat
(
    id       bigint not null,
    username varchar(100) default '',
    title    varchar(100) default '',
    c_type   varchar(20)  default '',
    created  integer      default 0,
    enable   boolean      default false,
    diesel   boolean      default true,
    lalafo   boolean      default true,
    house    boolean      default true,
    photo    boolean      default false,
    usd      varchar(100) default '0:0',
    kgs      varchar(100) default '0:0'
);
CREATE UNIQUE INDEX chat_chat_index
    on chat (id);

---- create above / drop below ----
DROP TABLE IF EXISTS chat;
