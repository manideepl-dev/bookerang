package dev.manideeplanka.bookerang.common;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import io.github.cdimascio.dotenv.Dotenv;
import org.flywaydb.core.Flyway;
import org.jdbi.v3.core.Jdbi;

public class DatabaseConfig {

    static final HikariDataSource ds = createDataSource();

    static {
        Flyway.configure()
                .dataSource(ds)
                .locations("classpath:db/migration")
                .load()
                .migrate();
    }

    public static Jdbi getJdbi() {
        return Jdbi.create(ds);
    }

    public static HikariDataSource createDataSource() {
        Dotenv dotenv = Dotenv.configure().ignoreIfMissing().load();
        HikariConfig config = new HikariConfig();
        config.setUsername(dotenv.get("DB_USERNAME"));
        config.setPassword(dotenv.get("DB_PASSWORD"));
        config.setJdbcUrl(dotenv.get("DB_URL"));

        return new HikariDataSource(config);
    }
}
