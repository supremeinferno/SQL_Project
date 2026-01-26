CREATE TABLE IF NOT EXISTS admins (
                                      id INTEGER PRIMARY KEY AUTOINCREMENT,
                                      username TEXT,
                                      password TEXT
);

INSERT INTO admins (username, password)
VALUES ('admin', 'admin123');

CREATE TABLE IF NOT EXISTS students (
                                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                                        name TEXT,
                                        roll_no TEXT,
                                        room_no TEXT,
                                        username TEXT,
                                        password TEXT
);

CREATE TABLE IF NOT EXISTS complaints (
                                          id INTEGER PRIMARY KEY AUTOINCREMENT,
                                          student_id INTEGER,
                                          title TEXT,
                                          description TEXT,
                                          status TEXT DEFAULT 'Pending',
                                          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                          FOREIGN KEY (student_id) REFERENCES students(id)
    );
