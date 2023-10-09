CREATE ROLE sonarqube_? WITH LOGIN PASSWORD 'CkO9l7&QxRz#VtYs';GRANT sonarqube_? TO postgres;
CREATE DATABASE sonarqube_part? WITH ENCODING 'UTF8' OWNER sonarqube_?;
GRANT ALL PRIVILEGES ON DATABASE sonarqube_part? TO sonarqube_?;