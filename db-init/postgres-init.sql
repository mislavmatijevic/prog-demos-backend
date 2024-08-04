delete from videos;
delete from subtopics;
delete from topics;

INSERT INTO topics (id, name) VALUES
    (1, 'Osnove programiranja uz C++')
ON CONFLICT (id) DO NOTHING;

INSERT INTO subtopics (id, id_topic, name) VALUES
    (1, 1, 'Osnove jezika C++'),
    (2, 1, 'Upravljanje podacima u nizovima'),
    (3, 1, 'Slogovi i Unije'),
    (4, 1, 'Pokazivači i Vezana Lista'),
    (5, 1, 'Funkcije'),
    (6, 1, 'Rekurzije'),
    (7, 1, 'Upis u Datoteke')
ON CONFLICT (id) DO NOTHING;

INSERT INTO videos (id, id_subtopic, identifier, name) VALUES
    (1, 1, 'RyL2MjxgVj0', 'Kako Napisati C++ Program'),
    (2, 1, 'RcFVMaGdSKM', 'C++ Varijable'),
    (3, 1, 'BqdPEeVPSB0', 'C++ Logika'),
    (4, 1, '3COiJ6b5sq4', 'C++ Petlje'),
    (5, 2, 'BSvFewITLv4', 'C++ Polja'),
    (6, 2, 'NYbzVk5vncM', 'C++ Sortiranja'),
    (7, 3, 'fTXnvSbAbWE', 'C++ Slogovi i Unije (1/3)'),
    (8, 3, 'WKoyPZxLOWM', 'C++ Slogovi i Unije (2/3)'),
    (9, 3, '-iLFn0Ttbbo', 'C++ Slogovi i Unije (3/3)'),
    (10, 4, 'oN-RparzioU', 'C++ Pokazivači i Vezana Lista (1/3)'),
    (11, 4, 'JOWopUG_U4I', 'C++ Pokazivači i Vezana Lista (2/3)'),
    (12, 4, 'F1sNQxkxfbg', 'C++ Pokazivači i Vezana Lista (3/3 dodatno o pokazivačima)'),
    (13, 5, 'KDI61ExnKZs', 'C++ Funkcije'),
    (15, 6, '8TI-NIByHR0', 'C++ Rekurzije'),
    (16, 7, '4k_sr8_v75s', 'C++ Tekstualne Datoteke'),
    (17, 7, 'SmnwRqHMLuw', 'C++ Binarne Datoteke')
ON CONFLICT (id) DO NOTHING;