import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { lastValueFrom } from "rxjs";

@Injectable({
    providedIn: "root"
})
export class loginService {
    constructor(
        private httpClient: HttpClient,
    ){

    }

    login(model: {username: string, password: string}){
        return lastValueFrom(this.httpClient.post<{token: string}>('/api/login', model))
    }
}