-- This file is part of GoforPomodoro.
--
-- GoforPomodoro is free software: you can redistribute it and/or modify
-- it under the terms of the GNU Affero General Public License as published by
-- the Free Software Foundation, either version 3 of the License, or
-- (at your option) any later version.
--
-- GoforPomodoro is distributed in the hope that it will be useful,
-- but WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
-- GNU Affero General Public License for more details.
--
-- You should have received a copy of the GNU Affero General Public License
-- along with GoforPomodoro.  If not, see <http://www.gnu.org/licenses/>.

DROP TABLE IF EXISTS chat_settings;

CREATE TABLE IF NOT EXISTS chat_settings(
    chat_id                       INTEGER NOT NULL PRIMARY KEY,

    default_sprint_duration_set   INTEGER,
    default_pomodoro_duration_set INTEGER,
    default_rest_duration_set     INTEGER,

    running_sprint_duration_set   INTEGER,
    running_pomodoro_duration_set INTEGER,
    running_rest_duration_set     INTEGER,

    running_sprint_duration       INTEGER,
    running_pomodoro_duration     INTEGER,
    running_rest_duration         INTEGER,

    running_end_next_sprint_ts    TIMESTAMP,
    running_end_next_rest_ts      TIMESTAMP,

    running_is_cancel             INTEGER, -- bool
    running_is_paused             INTEGER, -- bool
    running_is_rest               INTEGER, -- bool
    running_is_finished           INTEGER, -- bool

    autorun                       INTEGER, -- bool
    is_group                      INTEGER, -- bool
    subscribers                   TEXT, -- we use this to store de-normalized arrays (encoded)

    active                        INTEGER -- bool
);

CREATE INDEX ex1 ON chat_settings(active) WHERE active = 1;
