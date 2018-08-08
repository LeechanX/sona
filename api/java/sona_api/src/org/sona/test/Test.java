package org.sona.test;

import org.sona.api.SonaApi;

import java.util.ArrayList;

public class Test {
    public static void main(String[] args) throws Exception {
        SonaApi api = null;
        try {
            api = new SonaApi("lebron.james.info");
        } catch (Exception e) {
            return ;
        }
        while (true) {
            String value = api.Get("player", "number");
            System.out.println(value);
            value = api.Get("player", "team");
            System.out.println(value);
            ArrayList<String> list = api.GetList("friends", "list");
            for (String item: list) {
                System.out.println(item);
            }
            Thread.sleep(10000);
        }

    }
}

