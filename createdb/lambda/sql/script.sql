CREATE ROLE sonarqube_? WITH LOGIN PASSWORD 'PASSWD';GRANT sonarqube_? TO postgres;
CREATE DATABASE sonarqube_part? WITH ENCODING 'UTF8' OWNER sonarqube_?;
GRANT ALL PRIVILEGES ON DATABASE sonarqube_part? TO sonarqube_?;
GRANT ALL ON SCHEMA public TO sonarqube_?; 
