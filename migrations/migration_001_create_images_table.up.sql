CREATE TABLE images
(
    image_id     VARCHAR(36) PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    original_url VARCHAR(255) NOT NULL,
    url_512      VARCHAR(255),
    url_256      VARCHAR(255),
    url_16       VARCHAR(255)
);
