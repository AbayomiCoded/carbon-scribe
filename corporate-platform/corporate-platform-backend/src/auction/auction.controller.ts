import {
  Controller,
  Get,
  Post,
  Param,
  Body,
  UseGuards,
  Put,
} from '@nestjs/common';
import { AuctionService } from './auction.service';
import { CreateAuctionDto } from './dto/create-auction.dto';
import { PlaceBidDto } from './dto/place-bid.dto';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { JwtPayload } from '../auth/interfaces/jwt-payload.interface';
import { RateLimit, RateLimits } from '../rate-limit/rate-limit.decorator';

@Controller('api/v1/auctions')
@UseGuards(JwtAuthGuard)
export class AuctionController {
  constructor(private readonly auctionService: AuctionService) {}

  @Get()
  async getAuctions() {
    return this.auctionService.getAuctions();
  }

  @Get(':id')
  async getAuctionById(@Param('id') id: string) {
    return this.auctionService.getAuctionById(id);
  }

  @Post()
  async createAuction(@Body() dto: CreateAuctionDto) {
    return this.auctionService.createAuction(dto);
  }

  @Put(':id/start')
  async startAuction(@Param('id') id: string) {
    return this.auctionService.startAuction(id);
  }

  /**
   * Place a bid on an auction
   * Rate limited to 5 bids per minute per user per auction
   */
  @Post(':id/bids')
  @RateLimit(RateLimits.BIDDING)
  @RateLimit(RateLimits.GLOBAL_AUCTION_BIDDING)
  async placeBid(
    @Param('id') auctionId: string,
    @CurrentUser() user: JwtPayload,
    @Body() dto: PlaceBidDto,
  ) {
    return this.auctionService.placeBid(
      auctionId,
      user.sub,
      user.companyId,
      dto,
    );
  }

  @Get(':id/bids')
  async getAuctionBids(@Param('id') auctionId: string) {
    return this.auctionService.getAuctionBids(auctionId);
  }

  @Post(':id/settle')
  async settleAuction(@Param('id') auctionId: string) {
    return this.auctionService.settleAuction(auctionId);
  }
}
