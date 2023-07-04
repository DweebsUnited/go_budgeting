DROP TABLE IF EXISTS e_grp;
CREATE TABLE e_grp (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    sort INTEGER NOT NULL DEFAULT (0)
);

DROP TABLE IF EXISTS a;
CREATE TABLE a (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    hidden INTEGER NOT NULL DEFAULT (0),
    offbudget INTEGER NOT NULL DEFAULT (0),
    debt INTEGER NOT NULL DEFAULT (0),
    institution TEXT NOT NULL,
    name TEXT NOT NULL,
    class INTEGER NOT NULL
);

DROP TABLE IF EXISTS e;
CREATE TABLE e (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    groupID INTEGER REFERENCES e_grp(ID) NOT NULL,
    hidden INTEGER NOT NULL DEFAULT (0),
    debtAccount INTEGER REFERENCES a(ID) UNIQUE,
    name TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT (''),
    goalType INTEGER NOT NULL DEFAULT(0),
    goalAmt INTEGER NOT NULL DEFAULT(0),
    goalTgt INTEGER NOT NULL DEFAULT(0),
    sort INTEGER NOT NULL DEFAULT (999)
);

DROP TABLE IF EXISTS a_t;
CREATE TABLE a_t (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    accountID INTEGER REFERENCES a(ID) NOT NULL,
    type INTEGER NOT NULL,
    envelopeID INTEGER REFERENCES e(ID) NOT NULL,
    postDate INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    cleared INTEGER NOT NULL,
    memo TEXT NOT NULL
);

DROP INDEX IF EXISTS a_t_date;
DROP INDEX IF EXISTS a_t_aid;
DROP INDEX IF EXISTS a_t_eid;
CREATE INDEX a_t_date ON a_t (postDate);
CREATE INDEX a_t_aid ON a_t (accountID);
CREATE INDEX a_t_eid ON a_t (envelopeID);

DROP TABLE IF EXISTS e_t;
CREATE TABLE e_t (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    envelopeID INTEGER REFERENCES e(ID) NOT NULL,
    postDate INTEGER NOT NULL,
    amount INTEGER NOT NULL
);

DROP INDEX IF EXISTS e_t_date;
DROP INDEX IF EXISTS e_t_eid;
CREATE INDEX e_t_date ON e_t (postDate);
CREATE INDEX e_t_eid ON e_t (envelopeID);

DROP TABLE IF EXISTS a_chk;
CREATE TABLE a_chk (
    accountID INTEGER REFERENCES a(ID) NOT NULL,
    month INTEGER NOT NULL,
    bal INTEGER NOT NULL DEFAULT(0),
    "in" INTEGER NOT NULL DEFAULT(0),
    out INTEGER NOT NULL DEFAULT(0),
    cleared INTEGER NOT NULL DEFAULT(0),

    PRIMARY KEY(accountID, month)
);

DROP TABLE IF EXISTS e_chk;
CREATE TABLE e_chk (
    envelopeID INTEGER REFERENCES e(ID) NOT NULL,
    month INTEGER NOT NULL,
    bal INTEGER NOT NULL DEFAULT(0),
    "in" INTEGER NOT NULL DEFAULT(0),
    out INTEGER NOT NULL DEFAULT(0),

    PRIMARY KEY(envelopeID, month)
);

DROP TABLE IF EXISTS s_chk;
CREATE TABLE s_chk (
    month INTEGER PRIMARY KEY,
    float INTEGER NOT NULL DEFAULT(0),
    income INTEGER NOT NULL DEFAULT(0),
    expenses INTEGER NOT NULL DEFAULT(0),
    delta INTEGER NOT NULL DEFAULT(0),
    banked INTEGER NOT NULL DEFAULT(0),
    netWorth INTEGER NOT NULL DEFAULT(0)
);

DELETE FROM sqlite_sequence;
INSERT INTO sqlite_sequence (name, seq) VALUES ('a', 0);
INSERT INTO sqlite_sequence (name, seq) VALUES ('a_t', 0);
INSERT INTO sqlite_sequence (name, seq) VALUES ('e_grp', 0);
INSERT INTO sqlite_sequence (name, seq) VALUES ('e', 0);
INSERT INTO sqlite_sequence (name, seq) VALUES ('e_t', 0);

-- Default Envelope Groups
INSERT INTO e_grp (name,sort) VALUES ('Misc', 999);
INSERT INTO e_grp (name,sort) VALUES ('Monthly Bills', 100);
INSERT INTO e_grp (name,sort) VALUES ('Spending Money', 200);
INSERT INTO e_grp (name,sort) VALUES ('Preparation', 300);
INSERT INTO e_grp (name,sort) VALUES ('Events', 400);
INSERT INTO e_grp (name,sort) VALUES ('Debt', 500);

-- Default envelopes
INSERT INTO e (groupID, name, sort) VALUES (1, 'Buffer', 999);

INSERT INTO e_chk (envelopeID, month) VALUES(1, 0);

INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (2, 'Rent/Mortgage', 2, 230000, 100);
INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (2, 'Groceries', 2, 70000, 200);
INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (2, 'Pets', 2, 5000, 300);
INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (2, 'Gas', 2, 13000, 400);
INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (2, 'Phone', 2, 21000, 500);
INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (2, 'Electricity + Gas', 2, 15000, 600);
INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (2, 'Internet', 2, 6500, 700);

INSERT INTO e_chk (envelopeID, month) VALUES(2, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(3, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(4, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(5, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(6, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(7, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(8, 0);

INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (3, 'Spending Money', 2, 75000, 100);
INSERT INTO e (groupID, name, sort) VALUES (3, 'Allowances', 200);
INSERT INTO e (groupID, name, goalType, goalTgt, sort)
VALUES (3, 'Subscriptions', 2, 10000, 400);
INSERT INTO e (groupID, name, sort) VALUES (3, 'Weed', 500);
INSERT INTO e (groupID, name, sort) VALUES (3, 'Booze', 600);

INSERT INTO e_chk (envelopeID, month) VALUES(9, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(10, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(11, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(12, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(13, 0);

INSERT INTO e (groupID, name, sort) VALUES (4, 'Car Insurance', 100);
INSERT INTO e (groupID, name, sort) VALUES (4, 'Car Maintenance', 200);
INSERT INTO e (groupID, name, sort) VALUES (4, 'Health', 300);
INSERT INTO e (groupID, name, goalType, goalAmt, sort)
VALUES (4, 'Investment', 1, 50000, 400);
INSERT INTO e (groupID, name, goalType, goalAmt, goalTgt, sort)
VALUES (4, 'Gifts', 3, 5000, 100000, 500);
INSERT INTO e (groupID, name, goalType, goalAmt, sort)
VALUES (4, 'House Fund', 1, 50000, 600);
INSERT INTO e (groupID, name, goalType, goalAmt, goalTgt, sort)
VALUES (4, 'Vacation', 3, 25000, 1000000, 700);

INSERT INTO e_chk (envelopeID, month) VALUES(14, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(15, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(16, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(17, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(18, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(19, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(20, 0);

INSERT INTO e (groupID, name, sort) VALUES (6, 'Car Payment', 100);
INSERT INTO e (groupID, name, sort) VALUES (6, 'Student Loan Payment', 200);
INSERT INTO e (groupID, name, sort) VALUES (6, 'Personal Loan Payment', 300);

INSERT INTO e_chk (envelopeID, month) VALUES(21, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(22, 0);
INSERT INTO e_chk (envelopeID, month) VALUES(23, 0);