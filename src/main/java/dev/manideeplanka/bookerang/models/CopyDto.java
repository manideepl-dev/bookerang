package dev.manideeplanka.bookerang.models;


import lombok.Builder;
import lombok.Data;

//TODO add imageUrl later
@Data
@Builder
public class CopyDto {
    String copyId;
    String title;
    String author;
}

