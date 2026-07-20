package dev.manideeplanka.bookerang.controllers;

import dev.manideeplanka.bookerang.models.IdRes;
import dev.manideeplanka.bookerang.models.LoginReq;
import dev.manideeplanka.bookerang.models.SignupReq;
import dev.manideeplanka.bookerang.services.UserService;
import dev.manideeplanka.bookerang.models.TokenRes;
import io.javalin.http.Context;
import io.javalin.http.HttpStatus;

public class UserController {

    UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    public void login(Context ctx) throws Exception {
        LoginReq req = ctx.bodyAsClass(LoginReq.class);
        String token = userService.login(req);
        ctx.json(new TokenRes(token)).status(HttpStatus.OK);
    }

    public void profile(Context ctx) {
        ctx.json(new IdRes("casD","dqd")).status(HttpStatus.OK);
    }


    public void signup(Context ctx) throws Exception {
        SignupReq req = ctx.bodyAsClass(SignupReq.class);
        String token = userService.signup(req);
        ctx.json(new TokenRes(token)).status(HttpStatus.OK);
    }
}
