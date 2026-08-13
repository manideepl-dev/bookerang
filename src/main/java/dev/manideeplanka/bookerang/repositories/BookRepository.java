package dev.manideeplanka.bookerang.repositories;

import dev.manideeplanka.bookerang.common.DatabaseConfig;
import dev.manideeplanka.bookerang.models.AddBookResult;
import dev.manideeplanka.bookerang.models.CopyDto;
import dev.manideeplanka.bookerang.models.NearbyBookDto;
import org.jdbi.v3.core.Jdbi;
import java.util.List;
import java.util.UUID;

public class BookRepository {

    Jdbi jdbi;

    public BookRepository() {
        this.jdbi = DatabaseConfig.getJdbi();
    }

    public AddBookResult addBook(String author, String title, String owner) {
        return jdbi.inTransaction(handle -> {
            UUID authorId = handle.createQuery("""
                            INSERT into authors (name) VALUES (:name)
                            ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
                            RETURNING author_id
                            """)
                    .bind("name", author)
                    .mapTo(UUID.class)
                    .one();


            UUID bookId = handle.createQuery("""
                            INSERT into books (title, author_id) VALUES (:title, :authorId)
                            ON CONFLICT (title, author_id) DO UPDATE SET title = EXCLUDED.title
                            RETURNING book_id
                            """)
                    .bind("title", title)
                    .bind("authorId", authorId)
                    .mapTo(UUID.class)
                    .one();

            return handle.createQuery("""
                            INSERT into copies (book_id, owner_id) VALUES (:bookId, :owner)
                            ON CONFLICT (book_id, owner_id) DO UPDATE
                            SET owner_id = EXCLUDED.owner_id
                            RETURNING copy_id, xmax = 0 AS added
                            """)
                    .bind("bookId", bookId)
                    .bind("owner", owner)
                    .map((resultSet, context) -> new AddBookResult(
                            resultSet.getString("copy_id"),
                            resultSet.getBoolean("added")))
                    .one();
        });
    }

    public List<CopyDto> myBooks(String username) {
        String sql = """
                    SELECT c.copy_id, b.title, a.name
                    FROM copies c
                    JOIN books b ON c.book_id = b.book_id
                    JOIN authors a ON b.author_id = a.author_id
                    WHERE c.owner_id = :username;
                """;

        return jdbi.withHandle(handle -> {
            return handle.createQuery(sql).bind("username", username)
                    .map((rs, ctx) -> CopyDto.builder()
                            .copyId(rs.getObject("copy_id", UUID.class).toString())
                            .title(rs.getString("title"))
                            .author(rs.getString("name"))
                            .build())
                    .list();
        });
    }

    public List<NearbyBookDto> nearbyBooks(String username, long radius) {
        String sql = """
                    SELECT c.copy_id,
                           b.title,
                           a.name,
                           owner.username AS owner_username,
                           owner.first_name AS owner_first_name,
                           owner.last_name AS owner_last_name
                    FROM users requester
                    JOIN users owner ON owner.username <> requester.username
                    JOIN copies c ON c.owner_id = owner.username
                    JOIN books b ON c.book_id = b.book_id
                    JOIN authors a ON b.author_id = a.author_id
                    WHERE requester.username = :username
                      AND requester.location IS NOT NULL
                      AND owner.location IS NOT NULL
                      AND ST_DWithin(owner.location, requester.location, :radius * 1000)
                    ORDER BY ST_Distance(owner.location, requester.location), b.title;
                """;

        return jdbi.withHandle(handle -> handle.createQuery(sql)
                .bind("username", username)
                .bind("radius", radius)
                .map((rs, ctx) -> NearbyBookDto.builder()
                        .copyId(rs.getObject("copy_id", UUID.class).toString())
                        .title(rs.getString("title"))
                        .author(rs.getString("name"))
                        .ownerUsername(rs.getString("owner_username"))
                        .ownerFirstName(rs.getString("owner_first_name"))
                        .ownerLastName(rs.getString("owner_last_name"))
                        .build())
                .list());
    }

}
