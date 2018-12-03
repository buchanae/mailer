package model

var startupSql = `PRAGMA foreign_keys = ON;`

var schemaSql = `
create table if not exists message (
  id integer not null primary key autoincrement,
  mailbox_id integer not null references mailbox(id) on delete restrict on update restrict,

  size integer not null,

  seen integer not null,
  answered integer not null,
  flagged integer not null,
  deleted integer not null,
  draft integer not null,
  recent integer not null,

  created datetime not null,
  content blob not null
);

create table if not exists header (
  message_id integer not null references message(id) on delete cascade on update cascade,

  key text not null,
  value text not null
);

create index if not exists header_key_index on header (key);

create table if not exists mailbox (
  id integer not null primary key autoincrement,

  name text not null,

  unique (name collate nocase)
);
`