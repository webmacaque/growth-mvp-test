INSERT INTO shops (id, name)
VALUES (1, 'Demo Shop')
ON CONFLICT (id) DO NOTHING;

INSERT INTO orders (shop_id, number, total, customer_name, created_at)
VALUES
    (1, 'A-1001', 1590.00, 'Анна', NOW() - INTERVAL '6 days'),
    (1, 'A-1002', 2490.00, 'Борис', NOW() - INTERVAL '5 days'),
    (1, 'A-1003', 940.00, 'Вероника', NOW() - INTERVAL '4 days'),
ON CONFLICT DO NOTHING;
