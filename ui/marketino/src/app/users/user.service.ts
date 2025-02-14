import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";

@Injectable({
    providedIn: "root"
})
export class UserService {
    constructor(private httpClient: HttpClient){

    }

    create(model: any){
        return this.httpClient.post<{id: number}>("/api/users/create", model)
    }
}