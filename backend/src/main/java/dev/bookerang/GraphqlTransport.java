package dev.bookerang;

import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Bean;
import org.springframework.graphql.server.WebGraphQlInterceptor;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;
import org.springframework.web.servlet.config.annotation.CorsRegistry;

@Configuration
public class GraphqlTransport implements WebMvcConfigurer {
  @Bean BCryptPasswordEncoder passwordEncoder() { return new BCryptPasswordEncoder(); }
  @Bean WebGraphQlInterceptor authHeader() {
    return (request, chain) -> {
      String header = request.getHeaders().getFirst("Authorization");
      String token = header != null && header.startsWith("Bearer ") ? header.substring(7) : "";
      request.configureExecutionInput((input,builder) -> builder.graphQLContext(java.util.Map.of("token",token)).build());
      return chain.next(request);
    };
  }
  @Override public void addCorsMappings(CorsRegistry registry) {
    registry.addMapping("/graphql").allowedOrigins("http://localhost:3000").allowedMethods("POST","OPTIONS").allowedHeaders("Content-Type","Authorization");
  }
}
