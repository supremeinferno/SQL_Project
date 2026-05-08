-- Hostel management system - PL/SQL bits for Oracle.
-- Has the usual stuff: sequences, tables, views, a few triggers
-- and one package that wraps the procedures/functions we call
-- from the Go side.

-- Sequences (Oracle doesn't have AUTOINCREMENT, so this is how we fake it)
CREATE SEQUENCE admin_seq    START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
CREATE SEQUENCE student_seq  START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
CREATE SEQUENCE complaint_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
CREATE SEQUENCE audit_seq    START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;

-- Tables
CREATE TABLE admins (
    id       NUMBER DEFAULT admin_seq.NEXTVAL PRIMARY KEY,
    username VARCHAR2(100) NOT NULL,
    password VARCHAR2(100) NOT NULL
);

CREATE TABLE students (
    id       NUMBER DEFAULT student_seq.NEXTVAL PRIMARY KEY,
    name     VARCHAR2(100),
    roll_no  VARCHAR2(50) UNIQUE,
    room_no  VARCHAR2(20),
    username VARCHAR2(100),
    password VARCHAR2(100)
);

CREATE TABLE complaints (
    id          NUMBER DEFAULT complaint_seq.NEXTVAL PRIMARY KEY,
    student_id  NUMBER NOT NULL,
    title       VARCHAR2(200),
    description CLOB,
    status      VARCHAR2(20) DEFAULT 'Pending'
                    CHECK (status IN ('Pending', 'In Progress', 'Resolved')),
    created_at  TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT fk_complaint_student FOREIGN KEY (student_id) REFERENCES students(id)
);

CREATE TABLE complaint_audit (
    id           NUMBER DEFAULT audit_seq.NEXTVAL PRIMARY KEY,
    complaint_id NUMBER NOT NULL,
    old_status   VARCHAR2(20),
    new_status   VARCHAR2(20),
    changed_at   TIMESTAMP DEFAULT SYSTIMESTAMP
);

-- Views
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
LEFT JOIN students s ON c.student_id = s.id
ORDER BY c.created_at DESC;
/

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
/

-- Triggers

-- Drop a row in the audit table every time a complaint's status changes.
CREATE OR REPLACE TRIGGER trg_complaint_status_audit
AFTER UPDATE OF status ON complaints
FOR EACH ROW
WHEN (OLD.status != NEW.status)
BEGIN
    INSERT INTO complaint_audit (complaint_id, old_status, new_status, changed_at)
    VALUES (:OLD.id, :OLD.status, :NEW.status, SYSTIMESTAMP);
END;
/

-- If a student is being deleted, get rid of their complaints first
-- so we don't trip the foreign key.
CREATE OR REPLACE TRIGGER trg_cascade_delete_complaints
BEFORE DELETE ON students
FOR EACH ROW
BEGIN
    DELETE FROM complaints WHERE student_id = :OLD.id;
END;
/

-- Don't ever let admin id=1 get deleted. That's the bootstrap account
-- and losing it locks everyone out.
CREATE OR REPLACE TRIGGER trg_protect_main_admin
BEFORE DELETE ON admins
FOR EACH ROW
BEGIN
    IF :OLD.id = 1 THEN
        RAISE_APPLICATION_ERROR(-20001, 'Cannot delete the main admin account.');
    END IF;
END;
/

-- The CHECK constraint already covers this, but I want the trigger here
-- too just in case someone adds a new status path that bypasses it.
CREATE OR REPLACE TRIGGER trg_validate_insert_status
BEFORE INSERT ON complaints
FOR EACH ROW
BEGIN
    IF :NEW.status IS NOT NULL
       AND :NEW.status NOT IN ('Pending', 'In Progress', 'Resolved') THEN
        RAISE_APPLICATION_ERROR(-20002,
            'Invalid status "' || :NEW.status ||
            '". Allowed: Pending, In Progress, Resolved.');
    END IF;
END;
/

-- Same check, but for updates.
CREATE OR REPLACE TRIGGER trg_validate_update_status
BEFORE UPDATE OF status ON complaints
FOR EACH ROW
BEGIN
    IF :NEW.status NOT IN ('Pending', 'In Progress', 'Resolved') THEN
        RAISE_APPLICATION_ERROR(-20002,
            'Invalid status "' || :NEW.status ||
            '". Allowed: Pending, In Progress, Resolved.');
    END IF;
END;
/

-- Package spec - this is the public interface we call from Go.
CREATE OR REPLACE PACKAGE hostel_pkg AS

    -- Add a new student. Blows up if the roll number is already taken.
    PROCEDURE add_student(
        p_name     IN VARCHAR2,
        p_roll_no  IN VARCHAR2,
        p_room_no  IN VARCHAR2,
        p_username IN VARCHAR2,
        p_password IN VARCHAR2
    );

    -- Flip a complaint's status. Validates the new value first.
    PROCEDURE update_complaint_status(
        p_complaint_id IN NUMBER,
        p_new_status   IN VARCHAR2
    );

    -- How many complaints does this student have? Used by delete_student.
    FUNCTION get_complaint_count(
        p_student_id IN NUMBER
    ) RETURN NUMBER;

    -- Dump pending complaints to DBMS_OUTPUT. Mostly for debugging from SQL*Plus.
    PROCEDURE list_pending_complaints;

    -- Nuke a student plus everything they've ever complained about.
    PROCEDURE delete_student(
        p_student_id IN NUMBER
    );

END hostel_pkg;
/

-- And here's the actual implementation.
CREATE OR REPLACE PACKAGE BODY hostel_pkg AS

    PROCEDURE add_student(
        p_name     IN VARCHAR2,
        p_roll_no  IN VARCHAR2,
        p_room_no  IN VARCHAR2,
        p_username IN VARCHAR2,
        p_password IN VARCHAR2
    ) AS
        v_count          NUMBER;
        e_duplicate_roll EXCEPTION;
    BEGIN
        SELECT COUNT(*) INTO v_count
        FROM students
        WHERE roll_no = p_roll_no;

        IF v_count > 0 THEN
            RAISE e_duplicate_roll;
        END IF;

        INSERT INTO students (name, roll_no, room_no, username, password)
        VALUES (p_name, p_roll_no, p_room_no, p_username, p_password);

        COMMIT;
        DBMS_OUTPUT.PUT_LINE('Student "' || p_name || '" registered successfully.');

    EXCEPTION
        WHEN e_duplicate_roll THEN
            DBMS_OUTPUT.PUT_LINE('Error: Roll number "' || p_roll_no || '" already exists.');
            ROLLBACK;
        WHEN OTHERS THEN
            DBMS_OUTPUT.PUT_LINE('Unexpected error: ' || SQLERRM);
            ROLLBACK;
            RAISE;
    END add_student;

    PROCEDURE update_complaint_status(
        p_complaint_id IN NUMBER,
        p_new_status   IN VARCHAR2
    ) AS
        v_old_status     VARCHAR2(20);
        e_invalid_status EXCEPTION;
    BEGIN
        IF p_new_status NOT IN ('Pending', 'In Progress', 'Resolved') THEN
            RAISE e_invalid_status;
        END IF;

        SELECT status INTO v_old_status
        FROM complaints
        WHERE id = p_complaint_id;

        UPDATE complaints
        SET    status = p_new_status
        WHERE  id     = p_complaint_id;

        COMMIT;
        DBMS_OUTPUT.PUT_LINE(
            'Complaint #' || p_complaint_id ||
            ': ' || v_old_status || ' -> ' || p_new_status
        );

    EXCEPTION
        WHEN e_invalid_status THEN
            DBMS_OUTPUT.PUT_LINE('Error: "' || p_new_status || '" is not a valid status.');
            ROLLBACK;
        WHEN NO_DATA_FOUND THEN
            DBMS_OUTPUT.PUT_LINE('Error: Complaint #' || p_complaint_id || ' not found.');
            ROLLBACK;
        WHEN OTHERS THEN
            DBMS_OUTPUT.PUT_LINE('Unexpected error: ' || SQLERRM);
            ROLLBACK;
            RAISE;
    END update_complaint_status;

    FUNCTION get_complaint_count(
        p_student_id IN NUMBER
    ) RETURN NUMBER AS
        v_count NUMBER;
    BEGIN
        SELECT COUNT(*) INTO v_count
        FROM complaints
        WHERE student_id = p_student_id;
        RETURN v_count;
    EXCEPTION
        WHEN OTHERS THEN
            RETURN 0;
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
            IF c_pending%ISOPEN THEN CLOSE c_pending; END IF;
            DBMS_OUTPUT.PUT_LINE('Error: ' || SQLERRM);
            RAISE;
    END list_pending_complaints;

    PROCEDURE delete_student(
        p_student_id IN NUMBER
    ) AS
        v_name            VARCHAR2(100);
        v_complaint_count NUMBER;
    BEGIN
        SELECT name INTO v_name
        FROM   students
        WHERE  id = p_student_id;

        v_complaint_count := get_complaint_count(p_student_id);

        -- The trigger trg_cascade_delete_complaints would handle this on its own,
        -- but I'm leaving the explicit DELETE in so anyone reading this proc
        -- can see what's actually happening without having to dig through triggers.
        DELETE FROM complaints WHERE student_id = p_student_id;
        DELETE FROM students   WHERE id         = p_student_id;

        COMMIT;
        DBMS_OUTPUT.PUT_LINE(
            'Deleted "' || v_name || '" and ' ||
            v_complaint_count || ' complaint(s).'
        );

    EXCEPTION
        WHEN NO_DATA_FOUND THEN
            DBMS_OUTPUT.PUT_LINE('Error: Student #' || p_student_id || ' not found.');
            ROLLBACK;
        WHEN OTHERS THEN
            DBMS_OUTPUT.PUT_LINE('Unexpected error: ' || SQLERRM);
            ROLLBACK;
            RAISE;
    END delete_student;

END hostel_pkg;
/
