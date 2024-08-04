INSERT INTO topics (id, name) VALUES
    (1, 'Osnove programiranja uz C++')
ON CONFLICT (id) DO NOTHING;

INSERT INTO subtopics (id, id_topic, name) VALUES
    (1, 1, 'Osnove jezika C++'),
    (2, 1, 'C++ Polja'),
    (3, 1, 'C++ Sortiranja'),
    (4, 1, 'C++ Slogovi i Unije'),
    (5, 1, 'C++ Pokazivači i Vezana Lista'),
    (6, 1, 'C++ Funkcije'),
    (7, 1, 'C++ Rekurzije'),
    (8, 1, 'C++ Tekstualne Datoteke'),
    (9, 1, 'C++ Binarne Datoteke')
ON CONFLICT (id) DO NOTHING;

INSERT INTO videos (id, id_subtopic, identifier, name) VALUES
    (1, 1, 'RyL2MjxgVj0', 'Kako Napisati C++ Program'),
    (2, 1, 'RcFVMaGdSKM', 'C++ Varijable'),
    (3, 1, 'BqdPEeVPSB0', 'C++ Logika'),
    (4, 1, '3COiJ6b5sq4', 'C++ Petlje'),
    (5, 2, 'BSvFewITLv4', 'C++ Polja'),
    (6, 3, 'NYbzVk5vncM', 'C++ Sortiranja'),
    (7, 4, 'fTXnvSbAbWE', 'C++ Slogovi i Unije (1/3)'),
    (8, 4, 'WKoyPZxLOWM', 'C++ Slogovi i Unije (2/3)'),
    (9, 4, '-iLFn0Ttbbo', 'C++ Slogovi i Unije (3/3)'),
    (10, 5, 'oN-RparzioU', 'C++ Pokazivači i Vezana Lista (1/3)'),
    (11, 5, 'JOWopUG_U4I', 'C++ Pokazivači i Vezana Lista (2/3)'),
    (12, 5, 'F1sNQxkxfbg', 'C++ Pokazivači i Vezana Lista (3/3 dodatno o pokazivačima)'),
    (13, 6, 'KDI61ExnKZs', 'C++ Funkcije'),
    (15, 7, '8TI-NIByHR0', 'C++ Rekurzije'),
    (16, 8, '4k_sr8_v75s', 'C++ Tekstualne Datoteke'),
    (17, 9, 'SmnwRqHMLuw', 'C++ Binarne Datoteke')
ON CONFLICT (id) DO NOTHING;