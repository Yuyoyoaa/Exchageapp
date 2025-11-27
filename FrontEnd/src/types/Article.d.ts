export interface Article {
  ID: number;
  Title: string;
  Content: string;
  Preview: string;
  Cover?: string;
  LikesCount: number;
  ViewsCount: number;
  AuthorID: number;
  CategoryID?: number;
  Status: string;
  CreatedAt?: string;
  UpdatedAt?: string;
}

export interface Comment {
  ID: number;
  ArticleID: number;
  UserID: number;
  UserName: string;
  Content: string;
  ParentID?: number;
  CreatedAt?: string;
}

export interface Category {
  ID: number;
  Name: string;
}

export interface LikeResponse {
  message: string;
  likesCount: number;
}