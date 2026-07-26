package dev.manideeplanka.bookerang.services;

import dev.manideeplanka.bookerang.models.AddBookReq;
import dev.manideeplanka.bookerang.models.AddBookResult;
import dev.manideeplanka.bookerang.models.CopyDto;
import dev.manideeplanka.bookerang.repositories.BookRepository;
import lombok.extern.slf4j.Slf4j;
import java.util.List;

@Slf4j
public class BookService {


    BookRepository bookRepository;


    public BookService(BookRepository bookRepository) {
        this.bookRepository = bookRepository;
    }

    public AddBookResult addBook(AddBookReq req, String username) {
        return bookRepository.addBook(req.author(), req.title(), username);
    }

    public List<CopyDto> myBooks(String username) {
        return bookRepository.myBooks(username);
    }
}
