package dev.manideeplanka.bookerang.controllers;

import dev.manideeplanka.bookerang.models.*;
import dev.manideeplanka.bookerang.services.BookService;
import io.javalin.http.Context;
import io.javalin.http.HttpStatus;
import lombok.extern.slf4j.Slf4j;
import java.util.List;

@Slf4j
public class BookController {

    BookService bookService;

    public BookController(BookService bookService) {
        this.bookService = bookService;
    }

    public void addBook(Context ctx) {
        AddBookReq req = ctx.bodyAsClass(AddBookReq.class);
        String username = ctx.attribute("username");
        AddBookResult result = bookService.addBook(req, username);
        String message = result.added()
                ? "Book added to your collection"
                : "You already listed this book";
        ctx.json(new IdRes(message, result.copyId()))
                .status(result.added() ? HttpStatus.CREATED : HttpStatus.CONFLICT);
    }

    public void myBooks(Context ctx) {
        String username = ctx.attribute("username");
        List<CopyDto> copies = bookService.myBooks(username);
        ctx.json(new MyBooksRes(copies)).status(HttpStatus.OK);
    }

    public void nearbyBooks(Context ctx) {
        String username = ctx.attribute("username");
        long radius = ctx.queryParamAsClass("radius", Long.class).get();
        List<NearbyBookDto> copies = bookService.nearbyBooks(username, radius);
        ctx.json(new NearbyBooksRes(copies)).status(HttpStatus.OK);
    }
}
