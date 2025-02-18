import { Injectable } from "@angular/core";
import { BehaviorSubject } from "rxjs";

@Injectable({
    providedIn: 'root'
})
export class MarketinoTitleService {
    private dataSubject = new BehaviorSubject<{
        title: string,
        icon: string
    }>({title: '', icon: ''});
    data$ = this.dataSubject.asObservable();

    sendData(data: {
        title: string,
        icon: string
    }) {
        this.dataSubject.next(data);
    }
}