SET SERVEROUTPUT ON;

-- Hostel Management System schema for Oracle Database Free.
-- This script is safe to rerun: it drops existing demo objects first,
-- then recreates tables, views, triggers, package code, and seed data.

BEGIN
    EXECUTE IMMEDIATE 'DROP PACKAGE hostel_pkg';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -4043 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP VIEW student_complaint_summary';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -942 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP VIEW admin_complaint_view';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -942 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP TABLE complaint_audit PURGE';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -942 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP TABLE complaints PURGE';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -942 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP TABLE students PURGE';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -942 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP TABLE admins PURGE';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -942 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP SEQUENCE admin_seq';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -2289 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP SEQUENCE student_seq';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -2289 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP SEQUENCE complaint_seq';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -2289 THEN RAISE; END IF;
END;
/

BEGIN
    EXECUTE IMMEDIATE 'DROP SEQUENCE audit_seq';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -2289 THEN RAISE; END IF;
END;
/

CREATE SEQUENCE admin_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
CREATE SEQUENCE student_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
CREATE SEQUENCE complaint_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
CREATE SEQUENCE audit_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;

CREATE TABLE admins (
    id       NUMBER DEFAULT admin_seq.NEXTVAL PRIMARY KEY,
    username VARCHAR2(100) NOT NULL UNIQUE,
    password VARCHAR2(100) NOT NULL
);

CREATE TABLE students (
    id       NUMBER DEFAULT student_seq.NEXTVAL PRIMARY KEY,
    name     VARCHAR2(100) NOT NULL,
    roll_no  VARCHAR2(50)  NOT NULL UNIQUE,
    room_no  VARCHAR2(20)  NOT NULL,
    username VARCHAR2(100) NOT NULL UNIQUE,
    password VARCHAR2(100) NOT NULL
);

CREATE TABLE complaints (
    id          NUMBER DEFAULT complaint_seq.NEXTVAL PRIMARY KEY,
    student_id  NUMBER NOT NULL,
    title       VARCHAR2(200) NOT NULL,
    description CLOB NOT NULL,
    status      VARCHAR2(20) DEFAULT 'Pending' NOT NULL
                    CHECK (status IN ('Pending', 'In Progress', 'Resolved')),
    created_at  TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
    CONSTRAINT fk_complaint_student
        FOREIGN KEY (student_id) REFERENCES students(id)
);

-- complaint_id is intentionally not a foreign key, so audit history remains
-- readable even after a complaint/student is deleted from the demo system.
CREATE TABLE complaint_audit (
    id           NUMBER DEFAULT audit_seq.NEXTVAL PRIMARY KEY,
    complaint_id NUMBER NOT NULL,
    student_id   NUMBER,
    student_name VARCHAR2(100),
    old_status   VARCHAR2(20),
    new_status   VARCHAR2(20),
    changed_at   TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL
);

CREATE OR REPLACE VIEW admin_complaint_view AS
SELECT
    c.id          AS complaint_id,
    s.name        AS student_name,
    s.roll_no,
    s.room_no,
    c.title,
    c.description,
    c.status,
    c.created_at
FROM complaints c
JOIN students s ON c.student_id = s.id;

CREATE OR REPLACE VIEW student_complaint_summary AS
SELECT
    s.id   AS student_id,
    s.name,
    s.roll_no,
    s.room_no,
    COUNT(c.id)                                               AS total_complaints,
    SUM(CASE WHEN c.status = 'Pending'     THEN 1 ELSE 0 END) AS pending_count,
    SUM(CASE WHEN c.status = 'In Progress' THEN 1 ELSE 0 END) AS in_progress_count,
    SUM(CASE WHEN c.status = 'Resolved'    THEN 1 ELSE 0 END) AS resolved_count
FROM students s
LEFT JOIN complaints c ON s.id = c.student_id
GROUP BY s.id, s.name, s.roll_no, s.room_no;

CREATE OR REPLACE TRIGGER trg_complaint_status_audit
AFTER UPDATE OF status ON complaints
FOR EACH ROW
WHEN (OLD.status != NEW.status)
DECLARE
    v_student_name students.name%TYPE;
BEGIN
    SELECT name
    INTO   v_student_name
    FROM   students
    WHERE  id = :OLD.student_id;

    INSERT INTO complaint_audit (
        complaint_id,
        student_id,
        student_name,
        old_status,
        new_status,
        changed_at
    ) VALUES (
        :OLD.id,
        :OLD.student_id,
        v_student_name,
        :OLD.status,
        :NEW.status,
        SYSTIMESTAMP
    );
END;
/

CREATE OR REPLACE TRIGGER trg_cascade_delete_complaints
BEFORE DELETE ON students
FOR EACH ROW
BEGIN
    DELETE FROM complaints WHERE student_id = :OLD.id;
END;
/

CREATE OR REPLACE TRIGGER trg_protect_main_admin
BEFORE DELETE ON admins
FOR EACH ROW
BEGIN
    IF :OLD.id = 1 THEN
        RAISE_APPLICATION_ERROR(-20001, 'Cannot delete the main admin account.');
    END IF;
END;
/

CREATE OR REPLACE TRIGGER trg_validate_insert_status
BEFORE INSERT ON complaints
FOR EACH ROW
BEGIN
    IF :NEW.status IS NOT NULL
       AND :NEW.status NOT IN ('Pending', 'In Progress', 'Resolved') THEN
        RAISE_APPLICATION_ERROR(
            -20002,
            'Invalid status "' || :NEW.status || '". Allowed: Pending, In Progress, Resolved.'
        );
    END IF;
END;
/

CREATE OR REPLACE TRIGGER trg_validate_update_status
BEFORE UPDATE OF status ON complaints
FOR EACH ROW
BEGIN
    IF :NEW.status NOT IN ('Pending', 'In Progress', 'Resolved') THEN
        RAISE_APPLICATION_ERROR(
            -20002,
            'Invalid status "' || :NEW.status || '". Allowed: Pending, In Progress, Resolved.'
        );
    END IF;
END;
/

CREATE OR REPLACE PACKAGE hostel_pkg AS
    PROCEDURE add_student(
        p_name     IN VARCHAR2,
        p_roll_no  IN VARCHAR2,
        p_room_no  IN VARCHAR2,
        p_username IN VARCHAR2,
        p_password IN VARCHAR2
    );

    PROCEDURE update_complaint_status(
        p_complaint_id IN NUMBER,
        p_new_status   IN VARCHAR2
    );

    FUNCTION get_complaint_count(
        p_student_id IN NUMBER
    ) RETURN NUMBER;

    PROCEDURE list_pending_complaints;

    PROCEDURE delete_student(
        p_student_id IN NUMBER
    );
END hostel_pkg;
/

CREATE OR REPLACE PACKAGE BODY hostel_pkg AS
    PROCEDURE add_student(
        p_name     IN VARCHAR2,
        p_roll_no  IN VARCHAR2,
        p_room_no  IN VARCHAR2,
        p_username IN VARCHAR2,
        p_password IN VARCHAR2
    ) AS
        v_count             NUMBER;
        e_duplicate_student EXCEPTION;
    BEGIN
        SELECT COUNT(*)
        INTO   v_count
        FROM   students
        WHERE  roll_no = p_roll_no
           OR  username = p_username;

        IF v_count > 0 THEN
            RAISE e_duplicate_student;
        END IF;

        INSERT INTO students (name, roll_no, room_no, username, password)
        VALUES (p_name, p_roll_no, p_room_no, p_username, p_password);

        COMMIT;
        DBMS_OUTPUT.PUT_LINE('Student "' || p_name || '" registered successfully.');
    EXCEPTION
        WHEN e_duplicate_student THEN
            ROLLBACK;
            RAISE_APPLICATION_ERROR(-20003, 'Roll number or username already exists.');
        WHEN OTHERS THEN
            ROLLBACK;
            RAISE;
    END add_student;

    PROCEDURE update_complaint_status(
        p_complaint_id IN NUMBER,
        p_new_status   IN VARCHAR2
    ) AS
        e_invalid_status EXCEPTION;
    BEGIN
        IF p_new_status NOT IN ('Pending', 'In Progress', 'Resolved') THEN
            RAISE e_invalid_status;
        END IF;

        UPDATE complaints
        SET    status = p_new_status
        WHERE  id     = p_complaint_id;

        IF SQL%ROWCOUNT = 0 THEN
            RAISE_APPLICATION_ERROR(-20004, 'Complaint #' || p_complaint_id || ' not found.');
        END IF;

        COMMIT;
        DBMS_OUTPUT.PUT_LINE('Complaint #' || p_complaint_id || ' updated to ' || p_new_status || '.');
    EXCEPTION
        WHEN e_invalid_status THEN
            ROLLBACK;
            RAISE_APPLICATION_ERROR(-20002, 'Invalid complaint status.');
        WHEN OTHERS THEN
            ROLLBACK;
            RAISE;
    END update_complaint_status;

    FUNCTION get_complaint_count(
        p_student_id IN NUMBER
    ) RETURN NUMBER AS
        v_count NUMBER;
    BEGIN
        SELECT COUNT(*)
        INTO   v_count
        FROM   complaints
        WHERE  student_id = p_student_id;

        RETURN v_count;
    END get_complaint_count;

    PROCEDURE list_pending_complaints AS
        CURSOR c_pending IS
            SELECT c.id, s.name, s.roll_no, c.title, c.created_at
            FROM   complaints c
            JOIN   students   s ON c.student_id = s.id
            WHERE  c.status = 'Pending'
            ORDER  BY c.created_at ASC;

        v_rec   c_pending%ROWTYPE;
        v_total NUMBER := 0;
    BEGIN
        DBMS_OUTPUT.PUT_LINE('========== PENDING COMPLAINTS ==========');

        OPEN c_pending;
        LOOP
            FETCH c_pending INTO v_rec;
            EXIT WHEN c_pending%NOTFOUND;

            v_total := v_total + 1;
            DBMS_OUTPUT.PUT_LINE(
                'ID: '      || v_rec.id      || CHR(9) ||
                'Student: ' || v_rec.name    || CHR(9) ||
                'Roll: '    || v_rec.roll_no || CHR(9) ||
                'Title: '   || v_rec.title   || CHR(9) ||
                'Date: '    || TO_CHAR(v_rec.created_at, 'YYYY-MM-DD HH24:MI')
            );
        END LOOP;

        DBMS_OUTPUT.PUT_LINE('Total pending: ' || v_total);
        CLOSE c_pending;
    EXCEPTION
        WHEN OTHERS THEN
            IF c_pending%ISOPEN THEN
                CLOSE c_pending;
            END IF;
            RAISE;
    END list_pending_complaints;

    PROCEDURE delete_student(
        p_student_id IN NUMBER
    ) AS
        v_name            students.name%TYPE;
        v_complaint_count NUMBER;
    BEGIN
        SELECT name
        INTO   v_name
        FROM   students
        WHERE  id = p_student_id;

        v_complaint_count := get_complaint_count(p_student_id);

        DELETE FROM complaints WHERE student_id = p_student_id;
        DELETE FROM students   WHERE id         = p_student_id;

        COMMIT;
        DBMS_OUTPUT.PUT_LINE(
            'Deleted "' || v_name || '" and ' || v_complaint_count || ' complaint(s).'
        );
    EXCEPTION
        WHEN NO_DATA_FOUND THEN
            ROLLBACK;
            RAISE_APPLICATION_ERROR(-20005, 'Student #' || p_student_id || ' not found.');
        WHEN OTHERS THEN
            ROLLBACK;
            RAISE;
    END delete_student;
END hostel_pkg;
/

INSERT INTO admins (username, password) VALUES ('admin', '123');
COMMIT;

PROMPT Hostel Management System schema loaded successfully.
