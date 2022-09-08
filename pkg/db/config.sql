ALTER DATABASE dayrepsdb
CHARACTER SET = utf8
COLLATE = utf8_general_ci;

CREATE TABLE IF NOT EXISTS person ( -- Усі люди
telegram_id bigint UNSIGNED NOT NULL AUTO_INCREMENT,
person_name varchar(100) NOT NULL DEFAULT 'Anonymous',
state_id tinyint UNSIGNED DEFAULT 0 NOT NULL,
first_interaction datetime NOT NULL,
last_interaction datetime DEFAULT NULL,
timezone smallint(4) DEFAULT 0,
reminder_time time,
language_id smallint,
PRIMARY KEY (telegram_id)
);

CREATE TABLE IF NOT EXISTS dct_state (
state_id tinyint UNSIGNED NOT NULL,
state_name varchar(100) NOT NULL,
PRIMARY KEY (state_id)
);

INSERT IGNORE INTO dct_state (state_id, state_name) VALUES
(0, 'Regular'),
(1, 'Premium'),
(2, 'Donater');

ALTER TABLE dayrepsdb.person
ADD CONSTRAINT CS_person_state_id FOREIGN KEY (state_id)
REFERENCES dayrepsdb.dct_state (state_id);

CREATE TABLE IF NOT EXISTS reps (
reps_amount mediumint NOT NULL,
reps_date date NOT NULL,
reps_id mediumint UNSIGNED NOT NULL,
PRIMARY KEY (reps_date, reps_id)
);

CREATE TABLE IF NOT EXISTS dct_reps (
reps_id mediumint UNSIGNED NOT NULL UNIQUE AUTO_INCREMENT,
reps_name varchar(30) NOT NULL,
telegram_id bigint UNSIGNED NOT NULL,
order_value smallint NOT NULL DEFAULT 0,
daily_goal mediumint NOT NULL DEFAULT 0,
monthly_goal mediumint NOT NULL DEFAULT 0,
annual_goal mediumint NOT NULL DEFAULT 0,
PRIMARY KEY(reps_name, telegram_id)
);

ALTER TABLE dayrepsdb.dct_reps
ADD CONSTRAINT CS_reps_telegram_id FOREIGN KEY (telegram_id)
REFERENCES dayrepsdb.person (telegram_id);

ALTER TABLE dayrepsdb.reps
ADD CONSTRAINT CS_reps_reps_id FOREIGN KEY (reps_id)
REFERENCES dayrepsdb.dct_reps (reps_id);


CREATE TABLE dct_language (
language_id mediumint UNSIGNED,
language_code varchar(5) NOT NULL,
language_name varchar(50) NOT NULL,
PRIMARY KEY(language_id)
);


SELECT p.person_name, r.reps_amount, dr.reps_name, r.reps_date FROM reps r
LEFT JOIN dct_reps dr ON dr.reps_id = r.reps_id
LEFT JOIN person p ON p.telegram_id = dr.telegram_id
;