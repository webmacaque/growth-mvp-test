DELETE FROM orders
WHERE shop_id = 1
  AND number IN ('A-1001', 'A-1002', 'A-1003', 'A-1004', 'A-1005', 'A-1006', 'A-1007', 'A-1008');
DELETE FROM shops WHERE id = 1;
