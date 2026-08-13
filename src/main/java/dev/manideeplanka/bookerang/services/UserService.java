package dev.manideeplanka.bookerang.services;

import dev.manideeplanka.bookerang.common.JwtUtils;
import dev.manideeplanka.bookerang.exceptions.InvalidCredentialsException;
import dev.manideeplanka.bookerang.exceptions.UserAlreadyExistsException;
import dev.manideeplanka.bookerang.exceptions.UserNotFoundException;
import dev.manideeplanka.bookerang.models.LoginReq;
import dev.manideeplanka.bookerang.models.SignupReq;
import dev.manideeplanka.bookerang.models.User;
import dev.manideeplanka.bookerang.repositories.UserRepository;
import org.mindrot.jbcrypt.BCrypt;

import java.util.Optional;

public class UserService {

    UserRepository userRepository;


    public UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    public String login(LoginReq req) throws Exception {
        Optional<User> u = userRepository.findByUsername(req.username());

        if (u.isEmpty()) {
            throw new UserNotFoundException();
        }

        if (!BCrypt.checkpw(req.password(), u.get().getPassword())) {
            throw new InvalidCredentialsException();
        }

        return JwtUtils.createToken(u.get().getUsername());
    }


    public String signup(SignupReq req) throws Exception {
        Optional<User> u = userRepository.findByUsername(req.username());

        if (u.isPresent()) {
            throw new UserAlreadyExistsException();
        }

        User user = new User();
        user.setFirstName(req.firstName());
        user.setLastName(req.lastName());
        user.setUsername(req.username());
        user.setLatitude(req.latitude());
        user.setLongitude(req.longitude());
        user.setPassword(BCrypt.hashpw(req.password(), BCrypt.gensalt()));
        userRepository.signup(user);

        return JwtUtils.createToken(user.getUsername());
    }
}
