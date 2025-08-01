CREATE TABLE ping_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    status BOOLEAN NOT NULL,
    loss_rate FLOAT NOT NULL
);
