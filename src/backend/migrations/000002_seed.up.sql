INSERT INTO shops (id, name)
VALUES (1, 'Demo Shop')
ON CONFLICT (id) DO NOTHING;

INSERT INTO orders (shop_id, number, total, customer_name, created_at)
VALUES
    (1, 'A-1001', 1590.00, 'Анна', NOW() - INTERVAL '6 days'),
    (1, 'A-1002', 2490.00, 'Борис', NOW() - INTERVAL '5 days'),
    (1, 'A-1003', 940.00, 'Вероника', NOW() - INTERVAL '4 days'),
    (1, 'A-1004', 3190.00, 'Глеб', NOW() - INTERVAL '3 days'),
    (1, 'A-1005', 780.00, 'Дарья', NOW() - INTERVAL '2 days'),
    (1, 'A-1006', 4120.00, 'Егор', NOW() - INTERVAL '1 day'),
    (1, 'A-1007', 1290.00, 'Жанна', NOW() - INTERVAL '12 hours'),
    (1, 'A-1008', 560.00, 'Илья', NOW() - INTERVAL '2 hours')
ON CONFLICT DO NOTHING;
