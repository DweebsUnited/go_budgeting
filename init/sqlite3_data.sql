-- Default Envelope Groups
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