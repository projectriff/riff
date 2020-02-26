package io.projectriff.fatsapp;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class UpperController {

    @RequestMapping("/")
    public String upper(@RequestParam String input) {
        return input.toUpperCase();
    }
}
