package dev.manideeplanka.bookerang.models;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class NearbyBookDto {
    String copyId;
    String title;
    String author;
    String ownerUsername;
    String ownerFirstName;
    String ownerLastName;
}
