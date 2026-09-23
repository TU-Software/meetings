CREATE TABLE rooms (
    id   uuid NOT NULL DEFAULT uuidv7(),
    name text NOT NULL,

    CONSTRAINT rooms_pkey        PRIMARY KEY (id),
    CONSTRAINT rooms_name_unique UNIQUE      (name)
);

CREATE TABLE messages (
    id      uuid    NOT NULL DEFAULT uuidv7(),
    room_id uuid    NOT NULL,
    sender  text    NOT NULL,
    content text    NOT NULL,
    CONSTRAINT messages_pkey       PRIMARY KEY (id),
    CONSTRAINT messages_room_id_fk FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
);
