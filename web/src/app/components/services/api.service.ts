import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface DataItem {
  id: number;
  name: string;
  description: string;
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {

  // Replace with your Go backend API URL
  private apiUrl = 'http://localhost:8080/v1/';  // Example endpoint

  constructor(private http: HttpClient) {}

  // Method to fetch data from the API
  getData(): Observable<DataItem[]> {
    return this.http.get<DataItem[]>(this.apiUrl);
  }
}