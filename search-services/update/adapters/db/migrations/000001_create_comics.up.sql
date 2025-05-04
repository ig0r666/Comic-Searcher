CREATE TABLE comics (          
    comic_id INTEGER NOT NULL UNIQUE,      
    image_url TEXT,        
    keywords TEXT[]       
);