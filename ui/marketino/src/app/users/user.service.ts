import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { MarketinoTitleService } from "../services/marketino-title-service";
import { Title } from "@angular/platform-browser";

@Injectable({
    providedIn: "root"
})
export class UserService {
    constructor(
        private httpClient: HttpClient){

    }

    create(model: any){
        return this.httpClient.post<{id: number}>("/api/users/create", model)
    }

    delete(ids: number[]){
        return this.httpClient.post<{id: number}>("/api/users/delete", {ids})
    }
}