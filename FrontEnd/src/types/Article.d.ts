export interface Article {
  id: number;
  title: string;
  content: string;
  preview: string;
  cover?: string;
  likesCount: number;
  viewsCount: number;
  authorId: number;
  categoryId?: number;
  status: string;
  CreatedAt?: string; // JSON 返回大写 C，保持一致
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

export interface Favorite {
  ID: number;           // 收藏记录本身的ID
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  UserID: number;
  ArticleID: number;
  
  // 关键：必须包含 Article 字段，类型为您定义的 Article 接口
  Article: Article; 
}