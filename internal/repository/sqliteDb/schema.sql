CREATE TABLE IF NOT EXISTS downloads (
     id INTEGER PRIMARY KEY,
     url TEXT NOT NULL,
     queue TEXT,
     file_name TEXT,
     status TEXT,
     progress INTEGER DEFAULT 0,
     headers TEXT,
     content_length INTEGER,
     content_type TEXT,
     accept_ranges BOOLEAN,
     ranges_count INTEGER,
     ranges TEXT,
     created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS queues (
      name TEXT PRIMARY KEY,
      folder TEXT,
      max_dl INTEGER,
      speed TEXT,
      time_range TEXT
);