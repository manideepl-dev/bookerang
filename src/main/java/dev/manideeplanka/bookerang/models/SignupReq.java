package dev.manideeplanka.bookerang.models;

public record SignupReq(
        String username,
        String password,
        String firstName,
        String lastName,
        Double latitude,
        Double longitude
) {
}
